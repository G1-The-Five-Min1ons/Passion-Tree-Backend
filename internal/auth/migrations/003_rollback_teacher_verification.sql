IF OBJECT_ID('teacher_verification_requests', 'U') IS NOT NULL
BEGIN
    DROP TABLE teacher_verification_requests;
END
GO

IF COL_LENGTH('profile', 'Phone_Number') IS NOT NULL
BEGIN
    ALTER TABLE profile
    DROP COLUMN Phone_Number;
END
GO
