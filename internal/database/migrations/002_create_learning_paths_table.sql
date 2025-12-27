-- +migrate Up
-- Create LearningPaths table
CREATE TABLE LearningPaths (
    LearningPathID NVARCHAR(36) PRIMARY KEY DEFAULT NEWID(),
    UserID NVARCHAR(36) NOT NULL,
    Title NVARCHAR(255) NOT NULL,
    Description NVARCHAR(MAX) NULL,
    CreatedAt DATETIME2 NOT NULL DEFAULT GETDATE(),
    UpdatedAt DATETIME2 NOT NULL DEFAULT GETDATE(),
    CONSTRAINT FK_LearningPaths_Users FOREIGN KEY (UserID) REFERENCES Users(UserID) ON DELETE CASCADE
);

-- Create index on UserID for faster queries
CREATE INDEX IX_LearningPaths_UserID ON LearningPaths(UserID);

-- +migrate Down
DROP TABLE IF EXISTS LearningPaths;
