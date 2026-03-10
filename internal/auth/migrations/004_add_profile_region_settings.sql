-- Ensure phone number exists for account verification flow
IF COL_LENGTH('profile', 'Phone_Number') IS NULL
BEGIN
    ALTER TABLE profile
    ADD Phone_Number NVARCHAR(20) NULL;
END
GO

-- Add profile region settings defaults used by frontend
IF COL_LENGTH('profile', 'Time_Zone') IS NULL
BEGIN
    ALTER TABLE profile
    ADD Time_Zone NVARCHAR(64) NOT NULL
        CONSTRAINT DF_profile_time_zone DEFAULT 'Asia/Bangkok';
END
GO

IF COL_LENGTH('profile', 'Date_Format') IS NULL
BEGIN
    ALTER TABLE profile
    ADD Date_Format NVARCHAR(32) NOT NULL
        CONSTRAINT DF_profile_date_format DEFAULT 'DD/MM/YYYY';
END
GO

-- Ensure teacher verification table exists
IF OBJECT_ID('teacher_verification_requests', 'U') IS NULL
BEGIN
    CREATE TABLE teacher_verification_requests (
        request_id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT NEWID(),
        user_id UNIQUEIDENTIFIER NOT NULL,
        phone_number NVARCHAR(20) NOT NULL,
        reason NVARCHAR(500) NOT NULL,
        teaching_history NVARCHAR(MAX) NOT NULL,
        status NVARCHAR(20) NOT NULL DEFAULT 'pending',
        reviewed_by UNIQUEIDENTIFIER NULL,
        reviewed_at DATETIME NULL,
        create_at DATETIME NOT NULL DEFAULT GETDATE(),
        update_at DATETIME NOT NULL DEFAULT GETDATE(),

        CONSTRAINT FK_teacher_verification_user FOREIGN KEY (user_id) REFERENCES users(user_id),
        CONSTRAINT FK_teacher_verification_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(user_id),
        CONSTRAINT UQ_teacher_verification_user UNIQUE (user_id)
    );
END
GO