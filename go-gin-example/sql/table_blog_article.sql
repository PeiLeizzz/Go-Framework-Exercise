USE blog;
CREATE TABLE `blog_article` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `tag_id` int(10) unsigned DEFAULT '0' COMMENT '标签 ID',
    `title` varchar(100) DEFAULT '' COMMENT '文章标题',
    `desc` varchar(255) DEFAULT '' COMMENT '简述',
    `content` text,
    `create_on` int(11) DEFAULT NULL,
    `create_by` varchar(100) DEFAULT '' COMMENT '创建人',
    `modified_on` int(11) unsigned DEFAULT '0' COMMENT '修改时间',
    `modified_by` varchar(255) DEFAULT '' COMMENT '修改人',
    `delete_on` int(10) unsigned DEFAULT '0',
    `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态：0 为禁用、1 为启用',
    PRIMARY KEY(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章管理';