IF COL_LENGTH('profile', 'Date_Format') IS NOT NULL
BEGIN
    IF EXISTS (
        SELECT 1
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'Date_Format'
    )
    BEGIN
        DECLARE @df_date_format NVARCHAR(128)
        SELECT @df_date_format = dc.name
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'Date_Format'

        EXEC('ALTER TABLE profile DROP CONSTRAINT ' + @df_date_format)
    END

    ALTER TABLE profile
    DROP COLUMN Date_Format;
END
GO

IF COL_LENGTH('profile', 'Time_Zone') IS NOT NULL
BEGIN
    IF EXISTS (
        SELECT 1
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'Time_Zone'
    )
    BEGIN
        DECLARE @df_time_zone NVARCHAR(128)
        SELECT @df_time_zone = dc.name
        FROM sys.default_constraints dc
        INNER JOIN sys.columns c ON c.default_object_id = dc.object_id
        INNER JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.name = 'profile' AND c.name = 'Time_Zone'

        EXEC('ALTER TABLE profile DROP CONSTRAINT ' + @df_time_zone)
    END

    ALTER TABLE profile
    DROP COLUMN Time_Zone;
END
GO