CREATE TABLE public."Roles" (
    "ID"   INTEGER,
    "Name" TEXT NOT NULL UNIQUE,
    CONSTRAINT "Roles_pkey" PRIMARY KEY ("ID")
);

INSERT INTO public."Roles" ("ID", "Name") VALUES
    (1, 'user'),
    (2, 'moderator'),
    (3, 'admin');