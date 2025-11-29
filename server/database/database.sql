CREATE DATABASE IF NOT EXISTS note_sharing_application;
USE note_sharing_application;

CREATE TABLE Users (
    ID INT AUTO_INCREMENT,
    Username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci UNIQUE NOT NULL,
    PasswordHash VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    Salt VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL, -- Đã thêm cột Salt vào đây luôn
    PublicKey TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    PRIMARY KEY (ID)
);

CREATE TABLE Notes (
    ID INT AUTO_INCREMENT,
    OwnerID INT,
    Title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    EncryptedContent TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ID),
    FOREIGN KEY (OwnerID) REFERENCES Users(ID)
);