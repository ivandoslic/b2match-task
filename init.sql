-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION';

-- -----------------------------------------------------
-- Schema mydb
-- -----------------------------------------------------
-- -----------------------------------------------------
-- Schema event_management_solution
-- -----------------------------------------------------

-- -----------------------------------------------------
-- Schema event_management_solution
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `event_management_solution` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci ;
USE `event_management_solution` ;

-- -----------------------------------------------------
-- Table `event_management_solution`.`organizations`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`organizations` (
  `organization_id` INT NOT NULL AUTO_INCREMENT,
  `organization_name` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`organization_id`))
ENGINE = InnoDB
AUTO_INCREMENT = 17
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`users`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`users` (
  `user_id` INT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  `organization_id` INT NOT NULL,
  PRIMARY KEY (`user_id`),
  INDEX `organization_id` (`organization_id` ASC) VISIBLE,
  CONSTRAINT `users_ibfk_1`
    FOREIGN KEY (`organization_id`)
    REFERENCES `event_management_solution`.`organizations` (`organization_id`))
ENGINE = InnoDB
AUTO_INCREMENT = 11
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`authentication`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`authentication` (
  `user_id` INT NOT NULL,
  `username` VARCHAR(255) NOT NULL,
  `password` VARCHAR(255) NOT NULL,
  `auth_token` VARCHAR(255) NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  CONSTRAINT `authentication_ibfk_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`events`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`events` (
  `event_id` INT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `date` VARCHAR(10) NULL DEFAULT NULL,
  `organization_id` INT NOT NULL,
  `start_time` VARCHAR(8) NULL DEFAULT NULL,
  `end_time` VARCHAR(8) NULL DEFAULT NULL,
  PRIMARY KEY (`event_id`),
  INDEX `organization_id` (`organization_id` ASC) VISIBLE,
  CONSTRAINT `events_ibfk_1`
    FOREIGN KEY (`organization_id`)
    REFERENCES `event_management_solution`.`organizations` (`organization_id`))
ENGINE = InnoDB
AUTO_INCREMENT = 8
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`meetings`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`meetings` (
  `meeting_id` INT NOT NULL AUTO_INCREMENT,
  `event_id` INT NOT NULL,
  `scheduled_date` VARCHAR(255) NULL DEFAULT NULL,
  `scheduled_time` VARCHAR(255) NULL DEFAULT NULL,
  `organizer_id` INT NOT NULL,
  `duration` INT NOT NULL,
  PRIMARY KEY (`meeting_id`),
  INDEX `event_id` (`event_id` ASC) VISIBLE,
  INDEX `organizer_id` (`organizer_id` ASC) VISIBLE,
  CONSTRAINT `meetings_ibfk_1`
    FOREIGN KEY (`event_id`)
    REFERENCES `event_management_solution`.`events` (`event_id`),
  CONSTRAINT `meetings_ibfk_2`
    FOREIGN KEY (`organizer_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`))
ENGINE = InnoDB
AUTO_INCREMENT = 22
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`meeting_invitees`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`meeting_invitees` (
  `meeting_id` INT NOT NULL,
  `invitee_id` INT NOT NULL,
  `status` ENUM('Pending', 'Accepted', 'Rejected') NULL DEFAULT NULL,
  PRIMARY KEY (`meeting_id`, `invitee_id`),
  INDEX `invitee_id` (`invitee_id` ASC) VISIBLE,
  CONSTRAINT `meeting_invitees_ibfk_1`
    FOREIGN KEY (`meeting_id`)
    REFERENCES `event_management_solution`.`meetings` (`meeting_id`),
  CONSTRAINT `meeting_invitees_ibfk_2`
    FOREIGN KEY (`invitee_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`meeting_scheduling_status`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`meeting_scheduling_status` (
  `organizer_id` INT NULL DEFAULT NULL,
  `meeting_id` INT NULL DEFAULT NULL,
  `status` ENUM('Ready', 'Pending') NULL DEFAULT NULL,
  INDEX `organizer_id` (`organizer_id` ASC) VISIBLE,
  INDEX `meeting_id` (`meeting_id` ASC) VISIBLE,
  CONSTRAINT `meeting_scheduling_status_ibfk_1`
    FOREIGN KEY (`organizer_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`),
  CONSTRAINT `meeting_scheduling_status_ibfk_2`
    FOREIGN KEY (`meeting_id`)
    REFERENCES `event_management_solution`.`meetings` (`meeting_id`))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`user_events`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`user_events` (
  `user_event_id` INT NOT NULL AUTO_INCREMENT,
  `user_id` INT NULL DEFAULT NULL,
  `event_id` INT NULL DEFAULT NULL,
  PRIMARY KEY (`user_event_id`),
  INDEX `user_id` (`user_id` ASC) VISIBLE,
  INDEX `event_id` (`event_id` ASC) VISIBLE,
  CONSTRAINT `user_events_ibfk_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`),
  CONSTRAINT `user_events_ibfk_2`
    FOREIGN KEY (`event_id`)
    REFERENCES `event_management_solution`.`events` (`event_id`))
ENGINE = InnoDB
AUTO_INCREMENT = 23
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


-- -----------------------------------------------------
-- Table `event_management_solution`.`user_schedule`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `event_management_solution`.`user_schedule` (
  `user_id` INT NULL DEFAULT NULL,
  `event_id` INT NULL DEFAULT NULL,
  `start_time` TIME NULL DEFAULT NULL,
  `end_time` TIME NULL DEFAULT NULL,
  `meeting_id` INT NULL DEFAULT NULL,
  INDEX `user_id` (`user_id` ASC) VISIBLE,
  INDEX `event_id` (`event_id` ASC) VISIBLE,
  INDEX `meeting_id` (`meeting_id` ASC) VISIBLE,
  CONSTRAINT `user_schedule_ibfk_1`
    FOREIGN KEY (`user_id`)
    REFERENCES `event_management_solution`.`users` (`user_id`),
  CONSTRAINT `user_schedule_ibfk_2`
    FOREIGN KEY (`event_id`)
    REFERENCES `event_management_solution`.`events` (`event_id`),
  CONSTRAINT `user_schedule_ibfk_3`
    FOREIGN KEY (`meeting_id`)
    REFERENCES `event_management_solution`.`meetings` (`meeting_id`))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
