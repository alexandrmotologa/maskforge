--
-- MySQL database dump for MaskForge testing
--

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;

DROP TABLE IF EXISTS `customers`;
CREATE TABLE `customers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL,
  `full_name` varchar(150) DEFAULT NULL,
  `phone` varchar(50) DEFAULT NULL,
  `ssn` varchar(20) DEFAULT NULL,
  `credit_card` varchar(32) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `customers` (`id`, `email`, `full_name`, `phone`, `ssn`, `credit_card`) VALUES
(1, 'elena.popescu@fintech-cloud.ro', 'Elena Popescu', '+40721849201', '2850412400123', '4111111111111111'),
(2, 'james.wilson@enterprise.com', 'James Wilson', '+12025550184', '987-65-4321', '5500000000000004'),
(3, 'sophia.schmidt@berlin-tech.de', 'Sophia Schmidt', '+4915123456789', '543-21-0987', '378282246310005');
