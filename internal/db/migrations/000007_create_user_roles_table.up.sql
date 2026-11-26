CREATE TABLE public."UserRoles" (
    "UserID"    UUID,
    "RoleID"    INTEGER NOT NULL,
    CONSTRAINT "UserRoles_pkey" PRIMARY KEY ("UserID"),
    CONSTRAINT "FK_UserRoles_User" FOREIGN KEY ("UserID") REFERENCES public."Users"("ID"),
    CONSTRAINT "FK_UserRoles_Role" FOREIGN KEY ("RoleID") REFERENCES public."Roles"("ID")
);