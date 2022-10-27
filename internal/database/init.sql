DROP TABLE IF EXISTS "accounts";
DROP SEQUENCE IF EXISTS accounts_id_seq;
CREATE SEQUENCE accounts_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;
CREATE FUNCTION "update_modified_column" () RETURNS trigger LANGUAGE plpgsql AS '
BEGIN
    NEW.modified_at = now();
    RETURN NEW; 
END';

CREATE TABLE "accounts" (
    "id" integer DEFAULT nextval('accounts_id_seq') NOT NULL,
    "name" character varying NOT NULL,
    "email" character varying NOT NULL,
    "app_slug" character varying,
    "plan_slug" character varying,
    "resource_uuid" character varying NOT NULL,
    "language" character varying NOT NULL,
    "email_preference" boolean NOT NULL,
    "source" character varying,
    "source_id" character varying,
    "status" smallint NOT NULL,
    "license_key" character varying NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "modified_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "accounts_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "accounts_resource_uuid" UNIQUE ("resource_uuid")
) WITH (oids = false);


DELIMITER ;;

CREATE TRIGGER "accounts_bu" BEFORE UPDATE ON "accounts" FOR EACH ROW EXECUTE FUNCTION update_modified_column();;

DELIMITER ;

DROP TABLE IF EXISTS "activities";
DROP SEQUENCE IF EXISTS activities_id_seq;
CREATE SEQUENCE activities_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE "activities" (
    "id" integer DEFAULT nextval('activities_id_seq') NOT NULL,
    "account_id" integer NOT NULL,
    "resource_uuid" character varying NOT NULL,
    "type" character varying NOT NULL,
    "title" character varying NOT NULL,
    "body" character varying NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "modified_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "activities_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


DELIMITER ;;

CREATE TRIGGER "activities_bu" BEFORE UPDATE ON "activities" FOR EACH ROW EXECUTE FUNCTION update_modified_column();;

DELIMITER ;

DROP TABLE IF EXISTS "tokens";
DROP SEQUENCE IF EXISTS tokens_id_seq;
CREATE SEQUENCE tokens_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 32767 CACHE 1;

CREATE TABLE "tokens" (
    "id" smallint DEFAULT nextval('tokens_id_seq') NOT NULL,
    "resource_uuid" character varying NOT NULL,
    "access_token" character varying NOT NULL,
    "refresh_token" character varying NOT NULL,
    "expires_at" timestamptz NOT NULL,
    CONSTRAINT "tokens_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

CREATE INDEX "tokens_resource_uuid" ON "tokens" USING btree ("resource_uuid");
