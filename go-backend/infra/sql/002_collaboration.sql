-- F19: optional user public key for group key wrapping (PEM)
USE `smart_ledger`;

CREATE TABLE IF NOT EXISTS `user_public_keys` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `public_key_pem` TEXT NOT NULL,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  CONSTRAINT `fk_user_public_keys_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
