CREATE TABLE public."Tiles" (
    "TileId"   TEXT,
    "Rotation" INTEGER NOT NULL,
    "XPos"     INTEGER NOT NULL,
    "YPos"     INTEGER NOT NULL,
    "LayerId"  UUID    NOT NULL,
    CONSTRAINT "Tiles_LayerId_fkey" FOREIGN KEY ("LayerId") REFERENCES public."Layers" ("ID") ON DELETE CASCADE
);
