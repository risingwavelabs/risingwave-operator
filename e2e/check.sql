/* create a table */
create table t1(v1 int);

/* create a materialized view based on the previous table */
create materialized view mv1 as select sum(v1) as sum_v1 from t1;

/* insert some data into the source table */
insert into t1 values (1), (2), (3);

/* (optional) ensure the materialized view has been updated */
flush;

/* the materialized view should reflect the changes in source table */
select * from mv1;
