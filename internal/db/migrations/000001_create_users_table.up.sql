CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE public."Users" (
    "ID"        UUID,
    "Username"  TEXT NOT NULL UNIQUE,
    "Email"     CITEXT NOT NULL UNIQUE,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT "Users_pkey" PRIMARY KEY ("ID")
);
