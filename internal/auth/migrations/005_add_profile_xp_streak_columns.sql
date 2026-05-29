-- Add updated_at column used by AddXPAndRecalcLevel to stamp XP/level updates
IF COL_LENGTH('profile', 'updated_at') IS NULL
BEGIN
    ALTER TABLE profile
    ADD updated_at DATETIME NOT NULL
        CONSTRAINT DF_profile_updated_at DEFAULT GETDATE();
END
GO

-- Add Last_Active_Date column used by UpdateStreak to compute learning streaks
IF COL_LENGTH('profile', 'Last_Active_Date') IS NULL
BEGIN
    ALTER TABLE profile
    ADD Last_Active_Date DATE NULL;
END
GO
