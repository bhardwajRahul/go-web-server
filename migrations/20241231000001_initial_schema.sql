-- Create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "email" character varying(255) NOT NULL,
  "name" character varying(255) NOT NULL,
  "avatar_url" character varying(512) NULL,
  "bio" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);

-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");

-- Create index "idx_users_active" to table: "users"
CREATE INDEX "idx_users_active" ON "users" ("is_active");
