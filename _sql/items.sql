SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(4) ZEROFILL NOT NULL AUTO_INCREMENT PRIMARY KEY
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `features`;
CREATE TABLE `features` (
    `id` int(3)  NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `slug` VARCHAR(50) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `user_feature_relation`;
CREATE TABLE `user_feature_relation` (
    `userID` int(4) ZEROFILL NOT NULL,
    `featureID` int(3) NOT NULL,
    FOREIGN KEY (userID) REFERENCES users(id),
    FOREIGN KEY (featureID) REFERENCES features(id),
    UNIQUE (userID, featureID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`id`) VALUES
(1000),
(1001),
(1002),
(1003),
(1004),
(1005);

INSERT INTO `features` (`id`, `slug`) VALUES
(100, 'AVITO_VOICE_MESSAGES'),
(120, 'AVITO_PERFORMANCE_VAS'),
(256, 'AVITO_DISCOUNT_30'),
(588, 'AVITO_DISCOUNT_50');

INSERT INTO `user_feature_relation` (`userID`, `featureID`) VALUES
(1000, 100),
(1000, 120),
(1000, 256),
(1000, 588),
(1002, 100),
(1002, 588),
(1003, 256),
(1004, 588),
(1005, 256);
