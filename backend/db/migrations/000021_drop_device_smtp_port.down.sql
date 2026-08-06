alter table domain add column smtp_port int default null;
create unique index domain_smtp_port on domain (smtp_port);
