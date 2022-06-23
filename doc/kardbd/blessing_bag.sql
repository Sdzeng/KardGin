/*
 Navicat Premium Data Transfer

 Source Server         : 211.149.133.167
 Source Server Type    : MySQL
 Source Server Version : 80026
 Source Host           : 211.149.133.167:3306
 Source Schema         : blessing_bag

 Target Server Type    : MySQL
 Target Server Version : 80026
 File Encoding         : 65001

 Date: 21/04/2022 13:44:10
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for bb_bag
-- ----------------------------
DROP TABLE IF EXISTS `bb_bag`;
CREATE TABLE `bb_bag`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `bag_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `bag_sort` tinyint(0) NOT NULL COMMENT '排序',
  `namespace_id` int(0) NOT NULL,
  `status` tinyint(0) NOT NULL COMMENT '1-正常 2-删除',
  `create_time` datetime(0) NOT NULL,
  `update_time` datetime(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of bb_bag
-- ----------------------------
INSERT INTO `bb_bag` VALUES (1, '工具', 1, 1, 1, '2022-04-14 18:20:11', '2022-04-14 18:20:15');

-- ----------------------------
-- Table structure for bb_namespace
-- ----------------------------
DROP TABLE IF EXISTS `bb_namespace`;
CREATE TABLE `bb_namespace`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `namespace` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `status` tinyint(0) NOT NULL COMMENT '1-正常 2-删除',
  `create_time` datetime(0) NOT NULL,
  `update_time` datetime(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of bb_namespace
-- ----------------------------
INSERT INTO `bb_namespace` VALUES (1, 'scott', 1, '2022-04-14 18:18:52', '2022-04-14 18:18:54');

-- ----------------------------
-- Table structure for bb_treasure
-- ----------------------------
DROP TABLE IF EXISTS `bb_treasure`;
CREATE TABLE `bb_treasure`  (
  `id` int(0) NOT NULL AUTO_INCREMENT,
  `treasure_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `treasure_pic_url` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `treasure_link` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `click_cnt` int(0) NOT NULL DEFAULT 0,
  `bag_id` int(0) NOT NULL,
  `status` tinyint(0) NOT NULL COMMENT '1-正常 2-删除',
  `create_time` datetime(0) NOT NULL,
  `update_time` datetime(0) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of bb_treasure
-- ----------------------------
INSERT INTO `bb_treasure` VALUES (1, '电影台词查找器', 'great', 'great', 0, 1, 1, '2022-04-14 18:21:37', '2022-04-14 18:21:39');

SET FOREIGN_KEY_CHECKS = 1;
