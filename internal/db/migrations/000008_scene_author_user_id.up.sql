ALTER TABLE public."Scenes" DROP COLUMN "Author";
ALTER TABLE public."Scenes" ADD COLUMN "Author" UUID REFERENCES public."Users"("ID");
