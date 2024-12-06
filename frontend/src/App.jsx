import React, { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";

import { Grid } from "./Grid.jsx";
import { fetchBits } from "./utils/fetchBits.js";
import { fetchStats } from "./utils/fetchStats.js";
import { handleServerEvents } from "./utils/handleServerEvents.js";
import { BASE_WS_URL } from "./utils/constants.js";
import { Stats } from "./Stats.jsx";

const styles = {
  wrapper: {
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
  },

  header: {
    textAlign: "center",
    borderBottom: "solid 1px #000",
    width: "100%",
    padding: "5px",
    marginBottom: "20px",
  },
};

export const App = () => {
  const [bits, setBits] = useState(Array(1_000_000).fill(false));
  const [stats, setStats] = useState({ total: 0 });

  useEffect(() => {
    fetchBits(setBits);
    fetchStats(setStats);
  }, []);

  const WS_URL = `${BASE_WS_URL}/subscribe`;
  const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL, {
    share: false,
    shouldReconnect: (closeEvent) => {
      return closeEvent.code !== 1000;
    },
  });

  // Run when a new WebSocket message is received (lastMessage)
  useEffect(() => {
    handleServerEvents(lastMessage, setBits);
  }, [lastMessage]);

  // Checkbox event handler
  let checkboxToggleHandler = (e) => {
    if (readyState !== ReadyState.OPEN) {
      // @todo: Save event for retrying.
      console.log("Websocket connection is not open yet");
      return;
    }
    const checkbox = e.target;
    const id = parseInt(checkbox.id);

    setBits((prevState) => {
      const updatedState = [...prevState];
      updatedState[id] = checkbox.checked;
      return updatedState;
    });

    const action = checkbox.checked ? 1 : 0;
    const buffer = new ArrayBuffer(5);
    const view = new DataView(buffer);

    // Set the action byte
    // (0 for uncheck, 1 for check)
    view.setUint8(0, action);

    // Sent using big-endian byte order
    view.setUint32(1, id);

    sendMessage(buffer);
  };

  return (
    <div style={styles.wrapper}>
      <div style={styles.header}>
        <h1 style={{ margin: "0.5em" }}>One Million Checkboxes</h1>
        <Stats stats={stats} />
      </div>
      <Grid checkHandler={checkboxToggleHandler} bits={bits} />
    </div>
  );
};
