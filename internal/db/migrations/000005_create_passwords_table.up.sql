CREATE TABLE public."Passwords" (
    "ID"        UUID PRIMARY KEY,
    "PasswordHash"  TEXT NOT NULL,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "UpdatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT "FK_Passwords_User" FOREIGN KEY ("ID") REFERENCES public."Users"("ID")
);