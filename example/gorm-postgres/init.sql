-- DROP DATABASE IF EXISTS postgres;
-- CREATE DATABASE postgres;

\c postgres;

CREATE TABLE Student(StudentID SERIAL PRIMARY KEY, StudentName varchar(50));

INSERT INTO Student(StudentName)
VALUES ('Mary'),('Hannah');
