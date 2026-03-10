-- Drop trigger first if exists
IF OBJECT_ID('trg_settings_updated_at', 'TR') IS NOT NULL
BEGIN
    DROP TRIGGER trg_settings_updated_at;
END
GO

-- Drop settings table if exists
IF OBJECT_ID('settings', 'U') IS NOT NULL
BEGIN
    DROP TABLE settings;
END
GO
