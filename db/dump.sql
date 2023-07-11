--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2 (Debian 15.2-1.pgdg110+1)
-- Dumped by pg_dump version 15.2 (Debian 15.2-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: balance; Type: TABLE; Schema: public; Owner: dmitrij
--

CREATE TABLE public.balance (
    count integer,
    uuid integer
);


ALTER TABLE public.balance OWNER TO dmitrij;

--
-- Name: client; Type: TABLE; Schema: public; Owner: dmitrij
--

CREATE TABLE public.client (
    name character varying(50),
    uuid integer NOT NULL
);


ALTER TABLE public.client OWNER TO dmitrij;

--
-- Data for Name: balance; Type: TABLE DATA; Schema: public; Owner: dmitrij
--

COPY public.balance (count, uuid) FROM stdin;
11000	123
111000	115
\.


--
-- Data for Name: client; Type: TABLE DATA; Schema: public; Owner: dmitrij
--

COPY public.client (name, uuid) FROM stdin;
Андрей	115
Екатерина	123
\.


--
-- Name: client client_pk; Type: CONSTRAINT; Schema: public; Owner: dmitrij
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT client_pk PRIMARY KEY (uuid);


--
-- Name: balance balance_fk; Type: FK CONSTRAINT; Schema: public; Owner: dmitrij
--

ALTER TABLE ONLY public.balance
    ADD CONSTRAINT balance_fk FOREIGN KEY (uuid) REFERENCES public.client(uuid);


--
-- PostgreSQL database dump complete
--

