CREATE DATABASE note_sharing_application

USE note_sharing_application

CREATE TABLE Users (
    ID VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    Username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci UNIQUE,
    PasswordHash VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    PublicKey TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    PRIMARY KEY (ID)
);

CREATE TABLE Notes (
    ID VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    OwnerID VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    Title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    EncryptedContent TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ID),
    FOREIGN KEY (OwnerID) REFERENCES Users(ID)
);


-- Chen du lieu de test

INSERT INTO Users (ID, Username, PasswordHash, PublicKey) VALUES
('u1aaaaaa11', 'alice', 'hash_alice', 'alice_public_key'),
('u2bbbbbb22', 'bob', 'hash_bob', 'bob_public_key'),
('u3cccccc33', 'charlie', 'hash_charlie', 'charlie_public_key'),
('u4dddddd44', 'david', 'hash_david', 'david_public_key'),
('u5eeeeee55', 'eva', 'hash_eva', 'eva_public_key');

INSERT INTO Notes (ID, OwnerID, Title, EncryptedContent) VALUES
('n1aaaaaa11', 'u1aaaaaa11', 'Shopping list', 'encrypted_content_1'),
('n2bbbbbb22', 'u1aaaaaa11', 'Private note', 'encrypted_content_2'),
('n3cccccc33', 'u2bbbbbb22', 'Work plan', 'encrypted_content_3'),
('n4dddddd44', 'u3cccccc33', 'Secrets', 'encrypted_content_4'),
('n5eeeeee55', 'u4dddddd44', 'Crypto notes', 'encrypted_content_5');
