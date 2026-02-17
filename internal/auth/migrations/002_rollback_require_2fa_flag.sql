-- Rollback Migration: Remove require_2fa_next_login flag from Users table

-- Drop the index first
DROP INDEX IF EXISTS IX_Users_Require2FA ON Users;

-- Remove extended property (comment)
IF EXISTS (
    SELECT * FROM sys.extended_properties 
    WHERE major_id = OBJECT_ID('Users') 
    AND name = N'MS_Description' 
    AND minor_id = (SELECT column_id FROM sys.columns WHERE name = 'require_2fa_next_login' AND object_id = OBJECT_ID('Users'))
)
BEGIN
    EXEC sp_dropextendedproperty 
        @name = N'MS_Description',
        @level0type = N'SCHEMA', @level0name = N'dbo',
        @level1type = N'TABLE',  @level1name = N'Users',
        @level2type = N'COLUMN', @level2name = N'require_2fa_next_login';
END;

-- Drop Default Constraint before dropping the column
-- Some SQL Server versions require removing DEFAULT constraint before DROP COLUMN
DECLARE @ConstraintName nvarchar(200);
SELECT @ConstraintName = Name
FROM sys.default_constraints
WHERE parent_object_id = OBJECT_ID('Users')
AND parent_column_id = (SELECT column_id FROM sys.columns WHERE name = 'require_2fa_next_login' AND object_id = OBJECT_ID('Users'));

IF @ConstraintName IS NOT NULL
    EXEC('ALTER TABLE Users DROP CONSTRAINT ' + @ConstraintName);

-- Drop the column
ALTER TABLE Users
DROP COLUMN require_2fa_next_login;
