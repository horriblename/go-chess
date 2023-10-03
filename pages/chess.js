"use strict";

function dbg(x) {
  console.log(x);
  return x;
}

class AppState {
  constructor() {
    this.color = undefined;
    this.inTurn = false;
    // coordinate of the currently selected unit
    this.selected_unit = null;
    // coordinate of the target cell, to which to move to
    this.selected_target = null;
    this.socket = this.start_socket();
  }

  start_socket() {
    let socket = new WebSocket("ws://localhost:9990/game");

    socket.onopen = function (_e) {
      notify("Looking for opponent...");
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log(data);
      switch (data.message) {
        case "gameStart": {
          this.startGame(data.startFirst ? "white" : "black");
          document.getElementById("notification").innerHTML = "";
          this.inTurn = data.startFirst;
          break;
        }
        case "playerTurn": {
          if (data.check != "checkmate") {
            this.inTurn = true;
          }
          const [from, to] = data.opponentMove;
          this.move(this.coordFromNotation(from), this.coordFromNotation(to));
          break;
        }

        case "moveAccepted":
          this.move(this.selected_unit, this.selected_target);
          this.deselect_unit();
          this.selected_target = null;
          break;

        case "illegalMove": {
          this.deselect_unit();
          this.selected_target = null;
          this.inTurn = true;
          const notification = document.getElementById("notification");
          const notice = document.createElement("p", { class: "error" });
          notice.innerHTML = "Illegal Move";
          notification.replaceChildren(notice);
          break;
        }

        case "gameEnded":
          if (data.winner == "player") {
            notify("You won!");
          } else {
            notify("You lost!");
          }
          socket.close();
          break;

        default:
          console.log("received unknown websocket data ", data);
          break;
      }
    };

    socket.onclose = function (event) {
      if (event.wasClean) {
        alert(
          `[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`,
        );
      } else {
        // e.g. server process killed or network down
        // event.code is usually 1006 in this case
        alert("[close] Connection died");
      }
    };

    socket.onerror = function (_error) {
      alert(`[error]`);
    };
    return socket;
  }

  startGame(color) {
    if (color != "black" && color != "white") {
      console.trace("Error: invalid color");
      color = "white";
    }
    this.color = color;
    const opp_color = color == "white" ? "black" : "white";

    let home_row = [
      "rook",
      "knight",
      "bishop",
      "queen",
      "king",
      "bishop",
      "knight",
      "rook",
    ];
    let col_labels = ["a", "b", "c", "d", "e", "f", "g", "h"];
    let row_labels = [8, 7, 6, 5, 4, 3, 2, 1];
    if (color == "black") {
      home_row = home_row.reverse();
      col_labels = col_labels.reverse();
      row_labels = row_labels.reverse();
    }

    for (let i in home_row) {
      const type = home_row[i];
      document
        .getElementById(`cell-${i}-0`)
        .appendChild(AppState.newUnit(opp_color, type));
      document
        .getElementById(`cell-${i}-7`)
        .appendChild(AppState.newUnit(color, type));
    }

    for (let i = 0; i < 8; i++) {
      document
        .getElementById(`cell-${i}-1`)
        .appendChild(AppState.newUnit(opp_color, "pawn"));
      document
        .getElementById(`cell-${i}-6`)
        .appendChild(AppState.newUnit(color, "pawn"));
    }

    const board_body = document.getElementById("board-body");
    const board_header = document
      .getElementById("board-header")
      .getElementsByTagName("th");

    for (const i in col_labels) {
      board_header[i].appendChild(document.createTextNode(col_labels[i]));
    }

    const row_headers = board_body.getElementsByTagName("th");
    for (const i in row_labels) {
      row_headers[i].appendChild(document.createTextNode(row_labels[i]));
    }

    for (let i = 0; i < 8; i++) {
      for (let j = 0; j < 8; j++) {
        document.getElementById(`cell-${j}-${i}`).onclick = () =>
          this.select_cell(j, i);
      }
    }
  }

  static newUnit(color, type) {
    const unit = document.createElement("div");
    unit.classList.add(color, type);
    return unit;
  }

  move(from, to) {
    const cfrom = document.getElementById(`cell-${from.x}-${from.y}`);
    console.assert(cfrom);
    const cto = document.getElementById(`cell-${to.x}-${to.y}`);
    console.assert(cto);
    const unit = cfrom.removeChild(cfrom.firstChild);
    cto.replaceChildren(unit);
  }

  select_cell(x, y) {
    if (!this.inTurn) {
      return;
    }

    if (this.selected_unit == null) {
      if (this.selected_target) {
        // wait for server to reply whether move is accepted
        return;
      }

      const selection = document.getElementById(`cell-${x}-${y}`);
      console.assert(selection);
      if (
        selection.firstChild == null ||
        !selection.firstChild.classList.contains(this.color)
      ) {
        return;
      }
      this.select_unit(x, y);
    } else {
      const selection = document.getElementById(`cell-${x}-${y}`);

      const { x: x0, y: y0 } = this.selected_unit;
      const selected_unit = document.getElementById(`cell-${x0}-${y0}`);
      console.assert(selected_unit);

      // selection is another ally unit, change selected_unit to this instead
      if (
        selection.firstChild &&
        selection.firstChild.classList.contains(this.color)
      ) {
        this.deselect_unit();
        this.select_unit(x, y);
        return;
      }

      this.selected_target = { x: x, y: y };

      this.inTurn = false;
      this.socket.send(
        JSON.stringify({
          request: "move",
          move: [
            dbg(
              this.coordToNotation(this.selected_unit.x, this.selected_unit.y),
            ),
            dbg(this.coordToNotation(x, y)),
          ],
        }),
      );
    }
  }

  select_unit(x, y) {
    this.selected_unit = { x: x, y: y };
    const selection = document.getElementById(`cell-${x}-${y}`);
    selection.classList.add("selected");
  }

  deselect_unit() {
    const { x: x, y: y } = this.selected_unit;
    this.selected_unit = null;
    const selection = document.getElementById(`cell-${x}-${y}`);
    selection.classList.remove("selected");
  }

  coordToNotation(x, y) {
    if (this.color == "white") {
      y = 7 - y;
    } else {
      x = 7 - x;
    }

    return `${String.fromCodePoint("a".codePointAt(0) + x)}${y + 1}`;
  }

  coordFromNotation(s) {
    let x = s.codePointAt(0) - "a".codePointAt(0);
    let y = s.codePointAt(1) - "1".codePointAt(0);

    if (this.color == "white") {
      y = 7 - y;
    } else {
      x = 7 - x;
    }

    return { x: x, y: y };
  }
}

function notify(msg) {
  const notification = document.getElementById("notification");
  notification.innerHTML = msg;
}

window.onload = function () {
  const tbody = document.getElementById("board-body");

  for (let i = 0; i < 8; i++) {
    const tr = document.createElement("tr");
    tr.id = `row-${i}`;
    tr.appendChild(document.createElement("th"));

    for (let j = 0; j < 8; j++) {
      const td = document.createElement("td");
      td.id = `cell-${j}-${i}`;
      td.classList.add((i + j) % 2 == 0 ? "tile-dark" : "tile-light");
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }

  new AppState();
  // const socket = start_socket();
};
