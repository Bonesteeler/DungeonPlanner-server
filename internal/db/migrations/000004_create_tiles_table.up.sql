CREATE TABLE public."Tiles" (
    "ID"       UUID    NOT NULL,
    "TileId"   TEXT    NOT NULL,
    "Rotation" INTEGER NOT NULL,
    "XPos"     INTEGER NOT NULL,
    "YPos"     INTEGER NOT NULL,
    "LayerId"  UUID    NOT NULL,
    CONSTRAINT "Tiles_pkey" PRIMARY KEY ("ID"),
    CONSTRAINT "Tiles_LayerId_fkey" FOREIGN KEY ("LayerId") REFERENCES public."Layers" ("ID") ON DELETE CASCADE,
    CONSTRAINT "Tiles_LayerId_XPos_YPos_key" UNIQUE ("LayerId", "XPos", "YPos")
);
