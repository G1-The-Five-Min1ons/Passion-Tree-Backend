-- +migrate Up
-- Create Users table
CREATE TABLE Users (
    UserID NVARCHAR(36) PRIMARY KEY DEFAULT NEWID(),
    Email NVARCHAR(255) NOT NULL UNIQUE,
    PasswordHash NVARCHAR(255) NOT NULL,
    CreatedAt DATETIME2 NOT NULL DEFAULT GETDATE(),
    UpdatedAt DATETIME2 NOT NULL DEFAULT GETDATE()
);

-- Create index on Email for faster lookups
CREATE INDEX IX_Users_Email ON Users(Email);

-- +migrate Down
DROP TABLE IF EXISTS Users;
