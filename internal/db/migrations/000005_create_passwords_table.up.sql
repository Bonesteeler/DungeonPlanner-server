CREATE TABLE public."Passwords" (
    "ID"        UUID,
    "PasswordHash"  TEXT NOT NULL,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "UpdatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT "Passwords_pkey" PRIMARY KEY ("ID"),
    CONSTRAINT "FK_Passwords_User" FOREIGN KEY ("ID") REFERENCES public."Users"("ID")
);