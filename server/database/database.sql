CREATE DATABASE note_sharing_application

USE note_sharing_application

CREATE TABLE Users (
    ID INT AUTO_INCREMENT,
    Username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci UNIQUE,
    PasswordHash VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    PublicKey TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    PRIMARY KEY (ID)
);

CREATE TABLE Notes (
    ID INT AUTO_INCREMENT,
    OwnerID,
    Title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    EncryptedContent TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ID),
    FOREIGN KEY (OwnerID) REFERENCES Users(ID)
);


-- Chen du lieu de test

INSERT INTO Users (ID, Username, PasswordHash, PublicKey) VALUES
(1, 'alice', 'hash_alice', 'alice_public_key'),
(2, 'bob', 'hash_bob', 'bob_public_key'),
(3, 'charlie', 'hash_charlie', 'charlie_public_key'),
(4, 'david', 'hash_david', 'david_public_key'),
(5, 'eva', 'hash_eva', 'eva_public_key');

INSERT INTO Notes (ID, OwnerID, Title, EncryptedContent) VALUES
(1, 1, 'Shopping list', 'encrypted_content_1'),
(2, 2, 'Private note', 'encrypted_content_2'),
(3, 3, 'Work plan', 'encrypted_content_3'),
(4, 4, 'Secrets', 'encrypted_content_4'),
(5, 5, 'Crypto notes', 'encrypted_content_5');

-- drop database note_sharing_application


