CREATE TABLE public."Users" (
    "ID"        UUID,
    "Username"  TEXT NOT NULL UNIQUE,
    "Email"     TEXT NOT NULL UNIQUE,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT "Users_pkey" PRIMARY KEY ("ID")
);
