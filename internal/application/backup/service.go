package backupapp

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	inboundports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/inbound"
	outboundports "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/ports/outbound"
	"github.com/amirhossein-shakeri/zhinux-platform/apperr"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type Dependencies struct {
	IDGenerator        outboundports.BackupIDGenerator
	DatabaseRepository outboundports.DatabaseRepository
	JobRepository      outboundports.BackupJobRepository
	ArtifactRepository outboundports.BackupArtifactRepository
	DumpRunner         outboundports.DumpRunner
	Compressor         outboundports.Compressor
	ArtifactStore      outboundports.ArtifactStore
	ResourceProbe      outboundports.ResourceProbe
	EventPublisher     outboundports.BackupEventPublisher
	Clock              func() time.Time
}

type Service struct {
	deps Dependencies
}

func New(deps Dependencies) *Service {
	if deps.Clock == nil {
		deps.Clock = func() time.Time { return time.Now().UTC() }
	}
	return &Service{deps: deps}
}

func (s *Service) RequestBackup(ctx context.Context, command inboundports.RequestBackupCommand) (*inboundports.RequestBackupResult, error) {
	if command.DatabaseID <= 0 {
		return nil, apperr.New(apperr.CodeInvalidArgument, "database id is required")
	}
	if s.deps.IDGenerator == nil || s.deps.JobRepository == nil {
		return nil, apperr.New(apperr.CodeUnavailable, "backup request dependencies are not configured")
	}

	jobID, err := s.deps.IDGenerator.NewID(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "generate backup job id", err)
	}

	trigger := command.TriggerType
	if trigger == "" {
		trigger = backup.BackupTriggerManual
	}
	format := command.Format
	if format == "" {
		format = backup.BackupFormatPlainSQL
	}
	compression := command.Compression
	if compression.Algorithm == "" && compression.Mode == "" {
		compression = backup.DefaultZstdCompression()
	}

	job := &backup.BackupJob{
		ID:             jobID,
		DatabaseID:     command.DatabaseID,
		PlanID:         command.PlanID,
		TriggerType:    trigger,
		Status:         backup.BackupStatusPending,
		Format:         format,
		Compression:    compression,
		Actor:          command.Actor,
		CorrelationID:  command.CorrelationID,
		IdempotencyKey: command.IdempotencyKey,
	}

	if err := s.deps.JobRepository.Save(ctx, job); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "save backup job", err)
	}

	s.publish(ctx, backup.DomainEvent{
		EventName: backup.EventBackupRequested,
		At:        s.deps.Clock(),
		Data: backup.BackupRequestedPayload{
			JobID:         job.ID,
			DatabaseID:    job.DatabaseID,
			PlanID:        job.PlanID,
			TriggerType:   job.TriggerType,
			CorrelationID: job.CorrelationID,
		},
	})

	return &inboundports.RequestBackupResult{Job: job}, nil
}

func (s *Service) RunBackup(ctx context.Context, command inboundports.RunBackupCommand) (*inboundports.RunBackupResult, error) {
	if command.JobID <= 0 {
		return nil, apperr.New(apperr.CodeInvalidArgument, "backup job id is required")
	}
	if err := s.ensureRunDependencies(); err != nil {
		return nil, err
	}

	job, err := s.deps.JobRepository.FindByID(ctx, fmt.Sprintf("%d", command.JobID))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "find backup job", err)
	}
	if job == nil {
		return nil, apperr.New(apperr.CodeNotFound, "backup job not found")
	}

	sourceDatabase, err := s.deps.DatabaseRepository.FindByID(ctx, fmt.Sprintf("%d", job.DatabaseID))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "find source database", err)
	}
	if sourceDatabase == nil {
		return nil, apperr.New(apperr.CodeNotFound, "source database not found")
	}

	startedAt := s.deps.Clock()
	if !job.MarkStarted(startedAt) {
		return nil, apperr.New(apperr.CodeConflict, "backup job cannot be started from current status")
	}
	if err := s.deps.JobRepository.MarkStarted(ctx, fmt.Sprintf("%d", job.ID), startedAt); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "mark backup job started", err)
	}
	s.publish(ctx, backup.DomainEvent{
		EventName: backup.EventBackupStarted,
		At:        startedAt,
		Data:      backup.BackupStartedPayload{JobID: job.ID, StartedAt: startedAt},
	})

	artifactID, err := s.deps.IDGenerator.NewID(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "generate backup artifact id", err)
	}

	objectName := s.objectName(job)
	storeResult, compressionResult, dumpResult, err := s.streamBackup(ctx, streamRequest{
		jobID:          job.ID,
		artifactID:     artifactID,
		databaseID:     job.DatabaseID,
		objectName:     objectName,
		format:         job.Format,
		compression:    job.Compression,
		sourceDatabase: sourceDatabase,
	})
	dumpFinishedAt := s.deps.Clock()
	if err != nil {
		return nil, s.failJob(ctx, job, err)
	}

	resourceSnapshot := backup.ResourceSnapshot{}
	if s.deps.ResourceProbe != nil {
		if snapshot, probeErr := s.deps.ResourceProbe.Snapshot(ctx); probeErr == nil {
			resourceSnapshot = snapshot
		}
	}

	report := backup.BackupReport{
		DumpDuration:        dumpResult.Duration,
		CompressionDuration: compressionResult.Stats.Duration,
		TotalDuration:       dumpFinishedAt.Sub(startedAt),
		CompressionStats:    compressionResult.Stats,
		ResourceSnapshot:    resourceSnapshot,
	}

	artifact := &backup.BackupArtifact{
		ID:          artifactID,
		DatabaseID:  job.DatabaseID,
		BackupJobID: job.ID,
		Format:      job.Format,
		Storage:     storeResult.Storage,
		Size:        storeResult.Size,
		Checksum:    storeResult.Checksum,
		Compression: compressionResult.Stats,
		CreatedAt:   s.deps.Clock(),
	}
	if err := s.deps.ArtifactRepository.Save(ctx, artifact); err != nil {
		return nil, s.failJob(ctx, job, fmt.Errorf("save backup artifact: %w", err))
	}

	finishedAt := s.deps.Clock()
	if !job.MarkSucceeded(finishedAt, artifact.ID, report) {
		return nil, apperr.New(apperr.CodeConflict, "backup job cannot be completed from current status")
	}
	if err := s.deps.JobRepository.MarkFinished(ctx, fmt.Sprintf("%d", job.ID), job.Status, finishedAt, job.ArtifactID); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "mark backup job finished", err)
	}

	s.publish(ctx, backup.DomainEvent{
		EventName: backup.EventBackupArtifactCreated,
		At:        artifact.CreatedAt,
		Data: backup.BackupArtifactCreatedPayload{
			ArtifactID: artifact.ID,
			JobID:      job.ID,
			DatabaseID: job.DatabaseID,
			Storage:    artifact.Storage,
			Size:       artifact.Size,
		},
	}, backup.DomainEvent{
		EventName: backup.EventBackupCompleted,
		At:        finishedAt,
		Data:      backup.BackupCompletedPayload{JobID: job.ID, ArtifactID: artifact.ID, Report: report},
	})

	return &inboundports.RunBackupResult{Job: job, Artifact: artifact, Report: report}, nil
}

func (s *Service) ensureRunDependencies() error {
	if s.deps.IDGenerator == nil ||
		s.deps.DatabaseRepository == nil ||
		s.deps.JobRepository == nil ||
		s.deps.ArtifactRepository == nil ||
		s.deps.DumpRunner == nil ||
		s.deps.ArtifactStore == nil {
		return apperr.New(apperr.CodeUnavailable, "backup run dependencies are not configured")
	}
	return nil
}

type streamRequest struct {
	jobID          types.ID
	artifactID     types.ID
	databaseID     types.ID
	objectName     string
	format         backup.BackupFormat
	compression    backup.CompressionConfig
	sourceDatabase *database.Database
}

func (s *Service) streamBackup(ctx context.Context, request streamRequest) (*outboundports.PutArtifactResult, *outboundports.CompressionResult, *outboundports.DumpResult, error) {
	if request.sourceDatabase == nil {
		return nil, nil, nil, fmt.Errorf("source database is required")
	}

	if request.compression.Enabled {
		return s.streamCompressedBackup(ctx, request)
	}
	return s.streamPlainBackup(ctx, request)
}

func (s *Service) streamPlainBackup(ctx context.Context, request streamRequest) (*outboundports.PutArtifactResult, *outboundports.CompressionResult, *outboundports.DumpResult, error) {
	reader, writer := io.Pipe()
	storeCh := make(chan storeOutcome, 1)
	go func() {
		result, err := s.deps.ArtifactStore.Put(ctx, outboundports.PutArtifactRequest{
			JobID:       request.jobID,
			DatabaseID:  request.databaseID,
			ObjectName:  request.objectName,
			Content:     reader,
			ContentType: "application/sql",
		})
		storeCh <- storeOutcome{result: result, err: err}
	}()

	dumpResult, dumpErr := s.deps.DumpRunner.Dump(ctx, outboundports.DumpRequest{
		JobID:    request.jobID,
		Database: request.sourceDatabase,
		Format:   request.format,
		Writer:   writer,
	})
	_ = writer.CloseWithError(dumpErr)
	store := <-storeCh
	if dumpErr != nil {
		return nil, nil, dumpResult, dumpErr
	}
	if store.err != nil {
		return nil, nil, dumpResult, store.err
	}

	stats := backup.CompressionStats{
		Algorithm:         backup.CompressionAlgorithmNone,
		UncompressedBytes: store.result.Size,
		CompressedBytes:   store.result.Size,
	}
	return store.result, &outboundports.CompressionResult{Stats: stats}, dumpResult, nil
}

func (s *Service) streamCompressedBackup(ctx context.Context, request streamRequest) (*outboundports.PutArtifactResult, *outboundports.CompressionResult, *outboundports.DumpResult, error) {
	if s.deps.Compressor == nil {
		return nil, nil, nil, fmt.Errorf("compressor is not configured")
	}

	dumpReader, dumpWriter := io.Pipe()
	compressedReader, compressedWriter := io.Pipe()
	storeCh := make(chan storeOutcome, 1)
	compressCh := make(chan compressOutcome, 1)

	go func() {
		result, err := s.deps.ArtifactStore.Put(ctx, outboundports.PutArtifactRequest{
			JobID:       request.jobID,
			DatabaseID:  request.databaseID,
			ObjectName:  request.objectName,
			Content:     compressedReader,
			ContentType: "application/zstd",
		})
		storeCh <- storeOutcome{result: result, err: err}
	}()
	go func() {
		result, err := s.deps.Compressor.Compress(ctx, outboundports.CompressionRequest{
			JobID:  request.jobID,
			Config: request.compression,
			Source: dumpReader,
			Target: compressedWriter,
		})
		_ = compressedWriter.CloseWithError(err)
		compressCh <- compressOutcome{result: result, err: err}
	}()

	dumpResult, dumpErr := s.deps.DumpRunner.Dump(ctx, outboundports.DumpRequest{
		JobID:    request.jobID,
		Database: request.sourceDatabase,
		Format:   request.format,
		Writer:   dumpWriter,
	})
	_ = dumpWriter.CloseWithError(dumpErr)
	compress := <-compressCh
	store := <-storeCh

	if dumpErr != nil {
		return nil, compress.result, dumpResult, dumpErr
	}
	if compress.err != nil {
		return nil, compress.result, dumpResult, compress.err
	}
	if store.err != nil {
		return nil, compress.result, dumpResult, store.err
	}
	return store.result, compress.result, dumpResult, nil
}

type storeOutcome struct {
	result *outboundports.PutArtifactResult
	err    error
}

type compressOutcome struct {
	result *outboundports.CompressionResult
	err    error
}

func (s *Service) failJob(ctx context.Context, job *backup.BackupJob, cause error) error {
	finishedAt := s.deps.Clock()
	reason := cause.Error()
	_ = job.MarkFailed(finishedAt, reason)
	_ = s.deps.JobRepository.MarkFinished(ctx, fmt.Sprintf("%d", job.ID), job.Status, finishedAt, nil)
	s.publish(ctx, backup.DomainEvent{
		EventName: backup.EventBackupFailed,
		At:        finishedAt,
		Data:      backup.BackupFailedPayload{JobID: job.ID, Reason: reason},
	})
	return apperr.Wrap(apperr.CodeInternal, "run backup", cause)
}

func (s *Service) publish(ctx context.Context, events ...backup.DomainEvent) {
	if s.deps.EventPublisher == nil || len(events) == 0 {
		return
	}
	_ = s.deps.EventPublisher.Publish(ctx, events...)
}

func (s *Service) objectName(job *backup.BackupJob) string {
	extension := ".sql"
	if job.Format != backup.BackupFormatPlainSQL {
		extension = "." + strings.ReplaceAll(string(job.Format), "_", "-")
	}
	if job.Compression.Enabled && job.Compression.Algorithm == backup.CompressionAlgorithmZstd {
		extension += ".zst"
	}
	return fmt.Sprintf("database-%d/job-%d%s", job.DatabaseID, job.ID, extension)
}
