#!/bin/bash
read -r -d '' migration <<'EOF'
create table starts (
	id INTEGER PRIMARY KEY,
	game_id TEXT NOT NULL,
	json BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

create table moves (
	id INTEGER PRIMARY KEY,
	game_id TEXT NOT NULL,
	turn INTEGER NOT NULL,
	json BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

create table ends (
	id INTEGER PRIMARY KEY,
	game_id TEXT NOT NULL,
	json BLOB NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX starts_game_id ON starts(game_id);
CREATE INDEX moves_game_id ON moves(game_id);
CREATE INDEX moves_game_id_and_turn ON moves(game_id, turn);
CREATE INDEX ends_game_id ON ends(game_id);
EOF
sqlite3 ./games.sqlite "$migration"
