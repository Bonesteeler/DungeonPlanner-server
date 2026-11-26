CREATE TABLE public."Users" (
    "ID"        UUID PRIMARY KEY,
    "Username"  TEXT NOT NULL UNIQUE,
    "Email"     TEXT NOT NULL UNIQUE,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);
