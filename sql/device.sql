CREATE DATABASE IF NOT EXISTS device;

USE device;

CREATE TABLE  IF NOT EXISTS `domain_device_warehouse` (
  `sub_domain` varchar(32) NOT NULL,
  `device_id` varchar(32) NOT NULL,
  `device_type` int(8) NOT NULL DEFAULT '0',
  `public_key` varchar(64) DEFAULT NULL,
  `status` int(8) NOT NULL DEFAULT '1',
  `create_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  PRIMARY KEY (`sub_domain`, `device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `domain_device_mapping` (
  `did` bigint(20) NOT NULL AUTO_INCREMENT,
  `sub_domain` varchar(32) NOT NULL,
  `device_id` varchar(32) NOT NULL,
  `bind_token` varchar(32) DEFAULT NULL,
  `expire_time` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
   PRIMARY KEY (`did`),
   UNIQUE KEY (`sub_domain`, `device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `domain_device_info` (
  `did` bigint(20) NOT NULL,
  `hid` bigint(20) NOT NULL,
  `name` varchar(32) NOT NULL,
  `type` varchar(8) NOT NULL,
  `status` int(8) NOT NULL DEFAULT '1',
  `master_did` bigint(20) NOT NULL,
  `create_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  PRIMARY KEY (`did`),
  KEY (`hid`) USING HASH,
  KEY (`master_did`) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `domain_home_info` (
  `hid` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) DEFAULT 'Default',
  `status` int(8) NOT NULL DEFAULT '1',
  `create_uid` bigint(20) NOT NULL,
  `create_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  PRIMARY KEY (`hid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `domain_home_members` (
  `uid` bigint(20) NOT NULL,
  `hid` bigint(20) NOT NULL,
  `name` varchar(32) NOT NULL,
  `type` int(8) NOT NULL DEFAULT '1',
  `status` int(8) NOT NULL DEFAULT '1',
  `create_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  PRIMARY KEY (`uid`,`hid`),
  KEY (`hid`) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
