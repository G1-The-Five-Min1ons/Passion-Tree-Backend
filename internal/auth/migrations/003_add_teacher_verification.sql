-- Add phone number field for account binding
IF COL_LENGTH('profile', 'Phone_Number') IS NULL
BEGIN
    ALTER TABLE profile
    ADD Phone_Number NVARCHAR(20) NULL;
END
GO

-- Store teacher verification requests
IF OBJECT_ID('teacher_verification_requests', 'U') IS NULL
BEGIN
    CREATE TABLE teacher_verification_requests (
        request_id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY,
        user_id UNIQUEIDENTIFIER NOT NULL,
        phone_number NVARCHAR(20) NOT NULL,
        reason NVARCHAR(500) NOT NULL,
        teaching_history NVARCHAR(MAX) NOT NULL,
        status NVARCHAR(20) NOT NULL DEFAULT 'pending',
        reviewed_by UNIQUEIDENTIFIER NULL,
        reviewed_at DATETIME NULL,
        created_at DATETIME NOT NULL DEFAULT GETDATE(),
        updated_at DATETIME NOT NULL DEFAULT GETDATE(),

        CONSTRAINT FK_teacher_verification_user FOREIGN KEY (user_id) REFERENCES users(user_id),
        CONSTRAINT FK_teacher_verification_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(user_id),
        CONSTRAINT UQ_teacher_verification_user UNIQUE (user_id)
    );
END
GO
