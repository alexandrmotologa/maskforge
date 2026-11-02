--
-- PostgreSQL database dump for MaskForge testing
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(100),
    last_name character varying(100),
    phone_number character varying(50),
    ssn character varying(20),
    password_hash character varying(255),
    created_at timestamp without time zone DEFAULT now()
);

COPY public.users (id, email, first_name, last_name, phone_number, ssn, password_hash, created_at) FROM stdin;
1	alex.popescu@enterprise.ro	Alexandru	Popescu	+40722123456	1900101400123	$2a$12$K8y4X0OQ7aPq...	2026-01-10 08:30:00
2	sarah.connor@cyberdyne.com	Sarah	Connor	+12025550143	123-45-6789	$2a$12$L9z5Y1PR8bQr...	2026-01-11 09:15:00
3	michael.schumacher@ferrari.it	Michael	Schumacher	+49301234567	987-65-4321	$2a$12$M0a6Z2QS9cRs...	2026-01-12 11:20:00
4	jean.valjean@france-gov.fr	Jean	Valjean	+33123456789	543-21-0987	$2a$12$N1b7A3RT0dSt...	2026-01-13 14:40:00
5	ioana.radulescu@bank.ro	Ioana	Radulescu	+40744998877	2920202400567	$2a$12$O2c8B4SU1eTu...	2026-01-14 17:05:00
\.

CREATE TABLE public.payments (
    id integer NOT NULL,
    user_id integer NOT NULL,
    card_holder character varying(150),
    card_number character varying(32),
    iban character varying(34),
    billing_address character varying(255),
    ip_address character varying(45)
);

COPY public.payments (id, user_id, card_holder, card_number, iban, billing_address, ip_address) FROM stdin;
101	1	Alexandru Popescu	4111111111111111	RO49AAAA1B31007593840000	Strada Victoriei 10, Bucuresti	192.168.1.10
102	2	Sarah Connor	5500000000000004	GB29NWBK60161331926819	742 Evergreen Terrace, Springfield	10.0.4.15
103	3	Michael Schumacher	378282246310005	DE89370400440532013000	Unter den Linden 5, Berlin	198.51.100.4
104	4	Jean Valjean	4012888888881881	FR7630006000011234567890189	Rue Saint-Denis 15, Paris	192.0.2.77
105	5	Ioana Radulescu	5105105105105100	RO88BTRL00001234567890XX	Bulevardul Unirii 24, Bucuresti	172.16.8.20
\.

ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.payments ADD CONSTRAINT payments_pkey PRIMARY KEY (id);
