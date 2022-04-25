USE blog;
CREATE TABLE `blog_tag` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(100) DEFAULT '' COMMENT '标签名称',
    `create_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
    `create_by` varchar(100) DEFAULT '' COMMENT '创建人',
    `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
    `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
    `delete_on` int(10) unsigned DEFAULT '0' COMMENT '删除时间',
    `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态：0 为禁用、1 为启用',
    PRIMARY KEY(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章标签管理';