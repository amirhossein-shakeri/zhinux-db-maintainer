--
-- PostgreSQL database dump
--


-- Dumped from database version 15.3 (Debian 15.3-1.pgdg120+1)
-- Dumped by pg_dump version 18.4 (Ubuntu 18.4-1.pgdg24.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: backup_artifacts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.backup_artifacts (
    id text NOT NULL,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL,
    database_id bigint NOT NULL,
    backup_job_id text NOT NULL,
    storage_location text NOT NULL,
    size_bytes bigint NOT NULL,
    checksum text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT backup_artifacts_size_bytes_check CHECK ((size_bytes >= 0))
);


--
-- Name: backup_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.backup_jobs (
    id text NOT NULL,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL,
    database_id bigint NOT NULL,
    trigger_type text NOT NULL,
    status text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    artifact_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT backup_jobs_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'in_progress'::text, 'success'::text, 'failed'::text, 'canceled'::text]))),
    CONSTRAINT backup_jobs_trigger_type_check CHECK ((trigger_type = ANY (ARRAY['manual'::text, 'scheduled'::text])))
);


--
-- Name: backup_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.backup_plans (
    id text NOT NULL,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL,
    database_id bigint NOT NULL,
    schedule text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    retention_policy text NOT NULL,
    compression_enabled boolean DEFAULT false NOT NULL,
    encryption_enabled boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT backup_plans_retention_policy_check CHECK ((retention_policy = ANY (ARRAY['keep_last'::text, 'max_age'::text])))
);


--
-- Name: databases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.databases (
    id bigint NOT NULL,
    public_id uuid,
    title text NOT NULL,
    type text NOT NULL,
    host text NOT NULL,
    port integer NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT databases_port_check CHECK ((port > 0)),
    CONSTRAINT databases_type_check CHECK ((type = ANY (ARRAY['postgres'::text, 'mongo'::text])))
);


--
-- Name: databases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.databases ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.databases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: restore_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.restore_jobs (
    id text NOT NULL,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL,
    artifact_id text NOT NULL,
    target_database_id bigint NOT NULL,
    status text NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT restore_jobs_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'in_progress'::text, 'success'::text, 'failed'::text, 'canceled'::text])))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: backup_artifacts backup_artifacts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_artifacts
    ADD CONSTRAINT backup_artifacts_pkey PRIMARY KEY (id);


--
-- Name: backup_artifacts backup_artifacts_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_artifacts
    ADD CONSTRAINT backup_artifacts_public_id_key UNIQUE (public_id);


--
-- Name: backup_jobs backup_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_jobs
    ADD CONSTRAINT backup_jobs_pkey PRIMARY KEY (id);


--
-- Name: backup_jobs backup_jobs_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_jobs
    ADD CONSTRAINT backup_jobs_public_id_key UNIQUE (public_id);


--
-- Name: backup_plans backup_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_plans
    ADD CONSTRAINT backup_plans_pkey PRIMARY KEY (id);


--
-- Name: backup_plans backup_plans_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_plans
    ADD CONSTRAINT backup_plans_public_id_key UNIQUE (public_id);


--
-- Name: databases databases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.databases
    ADD CONSTRAINT databases_pkey PRIMARY KEY (id);


--
-- Name: databases databases_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.databases
    ADD CONSTRAINT databases_public_id_key UNIQUE (public_id);


--
-- Name: restore_jobs restore_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_pkey PRIMARY KEY (id);


--
-- Name: restore_jobs restore_jobs_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_public_id_key UNIQUE (public_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_backup_artifacts_backup_job_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_artifacts_backup_job_id ON public.backup_artifacts USING btree (backup_job_id);


--
-- Name: idx_backup_artifacts_database_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_artifacts_database_id ON public.backup_artifacts USING btree (database_id);


--
-- Name: idx_backup_artifacts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_artifacts_deleted_at ON public.backup_artifacts USING btree (deleted_at);


--
-- Name: idx_backup_artifacts_public_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_artifacts_public_id ON public.backup_artifacts USING btree (public_id);


--
-- Name: idx_backup_jobs_artifact_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_jobs_artifact_id ON public.backup_jobs USING btree (artifact_id);


--
-- Name: idx_backup_jobs_database_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_jobs_database_id ON public.backup_jobs USING btree (database_id);


--
-- Name: idx_backup_jobs_public_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_jobs_public_id ON public.backup_jobs USING btree (public_id);


--
-- Name: idx_backup_jobs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_jobs_status ON public.backup_jobs USING btree (status);


--
-- Name: idx_backup_plans_database_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_plans_database_id ON public.backup_plans USING btree (database_id);


--
-- Name: idx_backup_plans_public_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_backup_plans_public_id ON public.backup_plans USING btree (public_id);


--
-- Name: idx_databases_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_databases_deleted_at ON public.databases USING btree (deleted_at);


--
-- Name: idx_databases_host_port; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_databases_host_port ON public.databases USING btree (host, port);


--
-- Name: idx_databases_public_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_databases_public_id ON public.databases USING btree (public_id);


--
-- Name: idx_databases_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_databases_type ON public.databases USING btree (type);


--
-- Name: idx_restore_jobs_artifact_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_restore_jobs_artifact_id ON public.restore_jobs USING btree (artifact_id);


--
-- Name: idx_restore_jobs_public_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_restore_jobs_public_id ON public.restore_jobs USING btree (public_id);


--
-- Name: idx_restore_jobs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_restore_jobs_status ON public.restore_jobs USING btree (status);


--
-- Name: idx_restore_jobs_target_database_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_restore_jobs_target_database_id ON public.restore_jobs USING btree (target_database_id);


--
-- Name: backup_artifacts backup_artifacts_backup_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_artifacts
    ADD CONSTRAINT backup_artifacts_backup_job_id_fkey FOREIGN KEY (backup_job_id) REFERENCES public.backup_jobs(id) ON DELETE CASCADE;


--
-- Name: backup_artifacts backup_artifacts_database_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_artifacts
    ADD CONSTRAINT backup_artifacts_database_id_fkey FOREIGN KEY (database_id) REFERENCES public.databases(id) ON DELETE CASCADE;


--
-- Name: backup_jobs backup_jobs_database_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_jobs
    ADD CONSTRAINT backup_jobs_database_id_fkey FOREIGN KEY (database_id) REFERENCES public.databases(id) ON DELETE CASCADE;


--
-- Name: backup_plans backup_plans_database_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.backup_plans
    ADD CONSTRAINT backup_plans_database_id_fkey FOREIGN KEY (database_id) REFERENCES public.databases(id) ON DELETE CASCADE;


--
-- Name: restore_jobs restore_jobs_artifact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_artifact_id_fkey FOREIGN KEY (artifact_id) REFERENCES public.backup_artifacts(id) ON DELETE RESTRICT;


--
-- Name: restore_jobs restore_jobs_target_database_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_target_database_id_fkey FOREIGN KEY (target_database_id) REFERENCES public.databases(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


