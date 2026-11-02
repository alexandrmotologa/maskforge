export const DEMO_SQL_DUMP = `-- MaskForge Demonstration Database Dump (PostgreSQL & MySQL Compatible)
-- Includes multi-table relational schema with customer PII, payments, and orders.

CREATE TABLE users (
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
1001	elena.popescu@fintech-cloud.ro	Elena	Popescu	+40721849201	2850412400123	$2a$12$K8y4X0OQ7aPq...	2026-01-15 10:20:00
1002	james.wilson@enterprise.com	James	Wilson	+12025550184	987-65-4321	$2a$12$L9z5Y1PR8bQr...	2026-01-16 11:45:00
1003	sophia.schmidt@berlin-tech.de	Sophia	Schmidt	+4915123456789	543-21-0987	$2a$12$M0a6Z2QS9cRs...	2026-01-17 14:12:00
1004	adrian.ionescu@bucuresti.ro	Adrian	Ionescu	+40744991122	1900824400567	$2a$12$N1b7A3RT0dSt...	2026-01-18 09:05:00
1005	chloe.dupont@paris-solutions.fr	Chloe	Dupont	+33612345678	654-32-1098	$2a$12$O2c8B4SU1eTu...	2026-01-19 16:30:00
\\.

CREATE TABLE payment_methods (
    id integer NOT NULL,
    user_id integer NOT NULL,
    card_holder character varying(150),
    card_number character varying(32) NOT NULL,
    iban character varying(34),
    billing_address character varying(255),
    ip_address character varying(45)
);

COPY public.payment_methods (id, user_id, card_holder, card_number, iban, billing_address, ip_address) FROM stdin;
501	1001	Elena Popescu	4111111111111111	RO49AAAA1B31007593840000	Strada Mihai Eminescu 42, Bucuresti	192.168.1.45
502	1002	James Wilson	5500000000000004	GB29NWBK60161331926819	1200 Pennsylvania Ave NW, Washington DC	10.0.0.12
503	1003	Sophia Schmidt	378282246310005	DE89370400440532013000	Friedrichstrasse 108, Berlin	198.51.100.22
504	1004	Adrian Ionescu	4012888888881881	RO88BTRL00001234567890XX	Calea Victoriei 12, Bucuresti	172.16.4.88
505	1005	Chloe Dupont	5105105105105100	FR7630006000011234567890189	Rue de Rivoli 14, Paris	192.0.2.99
\\.

CREATE TABLE orders (
    order_id integer NOT NULL,
    customer_email character varying(255) NOT NULL,
    shipping_street character varying(255),
    shipping_city character varying(100),
    total_amount numeric(10,2),
    status character varying(50)
);

INSERT INTO orders (order_id, customer_email, shipping_street, shipping_city, total_amount, status) VALUES
(9001, 'elena.popescu@fintech-cloud.ro', 'Strada Mihai Eminescu 42', 'Bucuresti', 149.99, 'DELIVERED'),
(9002, 'james.wilson@enterprise.com', '1200 Pennsylvania Ave NW', 'Washington', 499.50, 'PROCESSING'),
(9003, 'sophia.schmidt@berlin-tech.de', 'Friedrichstrasse 108', 'Berlin', 89.00, 'SHIPPED');
`;
