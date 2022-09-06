create table syllabuses (
 id varchar(8) NOT NULL PRIMARY KEY,
 name varchar(50) NOT NULL,
 term varchar(20) NOT NULL
);

create table syllabus_relations (
  parent_id varchar(8) NOT NULL,
  child_id varchar(8) NOT NULL,
  PRIMARY KEY (parent_id, child_id)
);