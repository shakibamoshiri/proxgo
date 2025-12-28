package config

// DROP TABLE
const TABLE_DROP_USERS = `DROP TABLE users;`
const TABLE_DROP_BYTES = `DROP TABLE bytes;`
const TABLE_DROP_ARCHIVE = `DROP TABLE archive;`
const TABLE_DROP_FETCHED = `DROP TABLE fetched;`

// DELETE (wipe) TABLE
const TABLE_DELETE_USERS = `DELETE FROM users;`
const TABLE_DELETE_BYTES = `DELETE FROM bytes;`
const TABLE_DELETE_ARCHIVE = `DELETE FROM archive;`
const TABLE_DELETE_FETCHED = `DELETE FROM fetched;`

// CREATE TABLE
const TABLE_CREATE_USERS = `CREATE TABLE users (
    username  TEXT PRIMARY KEY,
    realname  TEXT NOT NULL,
    ctime     INTEGER NOT NULL,
    period    INTEGER NOT NULL,
    traffic   INTEGER NOT NULL,
    password  TEXT NOT NULL,
    page      TEXT NOT NULL,
    profile   TEXT NOT NULL
);`

const TABLE_CREATE_BYTES = `CREATE TABLE bytes (
    username    TEXT PRIMARY KEY,
    realname    TEXT NOT NULL,
    sessions INTEGER NOT NULL DEFAULT 0,
    ctime   INTEGER NOT NULL,
    atime   INTEGER NOT NULL,
    etime   INTEGER NOT NULL,

    bytes_base  INTEGER NOT NULL,
    bytes_used  INTEGER NOT NULL,
    bytes_pday  INTEGER NOT NULL,
    bytes_limit BOOLEAN NOT NULL,

    seconds_base  INTEGER NOT NULL,
    seconds_used  INTEGER NOT NULL,
    seconds_limit BOOLEAN NOT NULL
);`

const TABLE_CREATE_ARCHIVE = `CREATE TABLE archive (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL,
    realname      TEXT    NOT NULL,
    sessions      INTEGER NOT NULL DEFAULT 0,
    ctime         INTEGER NOT NULL,
    atime         INTEGER NOT NULL,
    etime         INTEGER NOT NULL,

    bytes_base    INTEGER NOT NULL,
    bytes_used    INTEGER NOT NULL,
    bytes_pday    INTEGER NOT NULL,
    bytes_limit   BOOLEAN NOT NULL,

    seconds_base  INTEGER NOT NULL,
    seconds_used  INTEGER NOT NULL,
    seconds_limit BOOLEAN NOT NULL
);`

const TABLE_CREATE_FETCHED = `CREATE TABLE fetched (
    username TEXT PRIMARY KEY,
    traffic  INTEGER NOT NULL,
    session  INTEGER NOT NULL
);`

const QUERY_USER_LIST = `
SELECT
    u.username,
    u.realname,
    COALESCE(f.traffic, 0),
    COALESCE(f.session, 0),
    CASE
        WHEN f.username IS NULL THEN 'unavailable'
        WHEN f.session = 0 THEN 'created'
        ELSE 'connected'
    END
FROM users u
LEFT JOIN fetched f ON u.username = f.username`

const QUERY_USER_STATS = `
SELECT
    b.username,
    b.realname,
    b.sessions,
    b.ctime,
    b.etime,
    b.bytes_used,
    b.bytes_pday,
    b.seconds_base,
    b.seconds_used
FROM bytes b;`

const QUERY_USER_SETUP = `
SELECT
    f.username,
    u.realname,
    f.session,
    u.traffic AS traffic_base,
    f.traffic AS traffic_used,
    u.period AS second_base,
    EXISTS (SELECT 1 FROM bytes b WHERE b.username = f.username) AS init
FROM fetched f
INNER JOIN users u ON f.username = u.username
WHERE f.session > 0;`
