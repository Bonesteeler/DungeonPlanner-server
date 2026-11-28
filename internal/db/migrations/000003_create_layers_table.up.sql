CREATE TABLE public."Layers" (
    "ID"      UUID    NOT NULL,
    "SceneId" UUID    NOT NULL,
    "Height"  INTEGER NOT NULL,
    CONSTRAINT "Layers_pkey" PRIMARY KEY ("ID"),
    CONSTRAINT "Layers_SceneId_fkey" FOREIGN KEY ("SceneId") REFERENCES public."Scenes" ("ID") ON DELETE CASCADE
);
