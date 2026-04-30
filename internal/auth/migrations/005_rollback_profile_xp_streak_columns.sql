IF COL_LENGTH('profile', 'Last_Active_Date') IS NOT NULL
BEGIN
    ALTER TABLE profile
    DROP COLUMN Last_Active_Date;
END
GO

IF COL_LENGTH('profile', 'updated_at') IS NOT NULL
BEGIN
    IF EXISTS (
        SELECT 1
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'updated_at'
    )
    BEGIN
        DECLARE @df_updated_at NVARCHAR(128)
        SELECT @df_updated_at = dc.name
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'updated_at'

        EXEC('ALTER TABLE profile DROP CONSTRAINT ' + @df_updated_at)
    END

    ALTER TABLE profile
    DROP COLUMN updated_at;
END
GO
