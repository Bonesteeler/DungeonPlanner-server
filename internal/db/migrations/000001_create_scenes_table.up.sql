CREATE TABLE public."Scenes" (
    "ID"               UUID        NOT NULL,
    "Name"             TEXT,
    "Author"           TEXT,
    "UniqueTileIDs"    TEXT[]      NOT NULL DEFAULT '{}',
    "ModerationStatus" INTEGER     NOT NULL DEFAULT 0,
    CONSTRAINT "Scenes_pkey" PRIMARY KEY ("ID")
);
