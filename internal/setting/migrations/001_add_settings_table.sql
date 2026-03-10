-- Create settings table for user-specific key-value preferences
IF OBJECT_ID('settings', 'U') IS NULL
BEGIN
    CREATE TABLE settings (
        id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT NEWID(),
        user_id UNIQUEIDENTIFIER NOT NULL,
        [key] NVARCHAR(100) NOT NULL,
        [value] NVARCHAR(MAX) NOT NULL,
        created_at DATETIME NOT NULL DEFAULT GETDATE(),
        updated_at DATETIME NOT NULL DEFAULT GETDATE(),

        CONSTRAINT FK_settings_user FOREIGN KEY (user_id) REFERENCES users(user_id),
        CONSTRAINT UQ_settings_user_key UNIQUE (user_id, [key])
    );
END
GO

-- Add indexes for common query patterns
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_id' AND object_id = OBJECT_ID('settings'))
BEGIN
    CREATE INDEX idx_settings_user_id ON settings(user_id);
END
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_settings_user_key' AND object_id = OBJECT_ID('settings'))
BEGIN
    CREATE INDEX idx_settings_user_key ON settings(user_id, [key]);
END
GO

-- Trigger to auto-update updated_at on row updates
IF OBJECT_ID('trg_settings_updated_at', 'TR') IS NULL
BEGIN
    EXEC(' 
        CREATE TRIGGER trg_settings_updated_at
        ON settings
        AFTER UPDATE
        AS
        BEGIN
            SET NOCOUNT ON;
            UPDATE s
            SET updated_at = GETDATE()
            FROM settings s
            INNER JOIN inserted i ON s.id = i.id;
        END
    ');
END
GO
