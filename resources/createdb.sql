CREATE TABLE uploads (
  id character(40) NOT NULL,
  filename character varying NOT NULL,
  title character varying,
  description character varying,
  "user" character varying NOT NULL,
  creation timestamp without time zone NOT NULL DEFAULT now(),
  expiration timestamp without time zone NOT NULL DEFAULT now(),
  size bigint NOT NULL,
  
  CONSTRAINT upload_size_check CHECK (size > 0),
  CONSTRAINT upload_time_check CHECK (creation < expiration)
);

CREATE INDEX uploads_expiration_idx
ON uploads
USING btree (
  expiration
);

CREATE INDEX uploads_id_idx
ON uploads
USING btree (
  id COLLATE pg_catalog."default"
);

CREATE INDEX uploads_id_size_idx
ON uploads
USING btree (
  id COLLATE pg_catalog."default",
  size
);

CREATE INDEX uploads_user_idx
ON uploads
USING btree (
  "user" COLLATE pg_catalog."default"
);
