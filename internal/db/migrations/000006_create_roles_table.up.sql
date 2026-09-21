CREATE TABLE public."Roles" (
    "ID"   INTEGER PRIMARY KEY,
    "Name" TEXT NOT NULL UNIQUE
);

INSERT INTO public."Roles" ("ID", "Name") VALUES
    (1, 'user'),
    (2, 'moderator'),
    (3, 'admin');