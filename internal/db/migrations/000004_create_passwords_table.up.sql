CREATE TABLE public."Passwords" (
    "ID"        UUID PRIMARY KEY,
    "PasswordHash"  TEXT NOT NULL,
    "Email"     TEXT NOT NULL,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "UpdatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);