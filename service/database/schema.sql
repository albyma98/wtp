-- Attivazione FK in SQLite
PRAGMA foreign_keys = ON;

-- Tabella user
CREATE TABLE user (
  uuid TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  photoUrl TEXT
);

-- Tabella conversation
CREATE TABLE conversation (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  isDirect BOOLEAN NOT NULL,
  groupName TEXT,
  groupPhoto TEXT,
  timestampCreated TEXT NOT NULL,
  timestampLastMessage TEXT NOT NULL
);

-- Tabella member
CREATE TABLE member (
  uuidUser TEXT NOT NULL,
  idConversation INTEGER NOT NULL,
  timestampJoined TEXT NOT NULL,
  PRIMARY KEY (uuidUser, idConversation),
  FOREIGN KEY (uuidUser) REFERENCES user(uuid) ON DELETE CASCADE,
  FOREIGN KEY (idConversation) REFERENCES conversation(id) ON DELETE CASCADE
);

-- Tabella message
CREATE TABLE message (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL CHECK(type IN ('text', 'photo')),
  content TEXT,
  mediaUrl TEXT,
  timestamp TEXT NOT NULL,
  idRepliesTo INTEGER,
  idForwardedFrom INTEGER,
  idConversation INTEGER NOT NULL,
  uuidSender TEXT,
  FOREIGN KEY (idRepliesTo) REFERENCES message(id) ON DELETE SET NULL,
  FOREIGN KEY (idForwardedFrom) REFERENCES message(id) ON DELETE SET NULL,
  FOREIGN KEY (idConversation) REFERENCES conversation(id) ON DELETE CASCADE,
  FOREIGN KEY (uuidSender) REFERENCES user(uuid) ON DELETE SET NULL
);

-- Tabella reaction
CREATE TABLE reaction (
  uuidUser TEXT NOT NULL,
  idMessage INTEGER NOT NULL,
  emoji TEXT NOT NULL,
  PRIMARY KEY (uuidUser, idMessage),
  FOREIGN KEY (uuidUser) REFERENCES user(uuid) ON DELETE CASCADE,
  FOREIGN KEY (idMessage) REFERENCES message(id) ON DELETE CASCADE
);

-- Tabella messageStatus
CREATE TABLE messageStatus (
  uuidUser TEXT NOT NULL,
  idMessage INTEGER NOT NULL,
  delivered BOOLEAN NOT NULL,
  seen BOOLEAN NOT NULL,
  PRIMARY KEY (uuidUser, idMessage),
  FOREIGN KEY (uuidUser) REFERENCES user(uuid) ON DELETE CASCADE,
  FOREIGN KEY (idMessage) REFERENCES message(id) ON DELETE CASCADE
);
