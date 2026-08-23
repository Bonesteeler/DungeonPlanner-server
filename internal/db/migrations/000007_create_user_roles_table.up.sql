CREATE TABLE public."UserRoles" (
    "UserID"    UUID PRIMARY KEY,
    "RoleID"    INTEGER NOT NULL,
    CONSTRAINT "FK_UserRoles_User" FOREIGN KEY ("UserID") REFERENCES public."Users"("ID"),
    CONSTRAINT "FK_UserRoles_Role" FOREIGN KEY ("RoleID") REFERENCES public."Roles"("ID")
);