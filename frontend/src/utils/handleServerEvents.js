export const handleServerEvents = (message, setState) => {
  if (message === null) {
    return;
  }

  const blob = message.data;

  if (!(blob instanceof Blob)) {
    console.error("Invalid message: Expected Blob but received something else");
    return;
  }

  blob
    .arrayBuffer()
    .then((buffer) => {
      // Create a DataView to read the binary data
      const view = new DataView(buffer);

      // Validate the message length
      if (view.byteLength % 5 !== 0) {
        console.error(
          `Invalid message: Expected sets of 5 bytes, but got ${view.byteLength}`
        );
        return;
      }
      setState((prevState) => {
        const updatedState = [...prevState];
        for (let i = 0; i < view.byteLength / 5; i++) {
          // Read the first byte as the action
          const action = view.getUint8(i * 5); // 0 for "uncheck", 1 for "check"

          // Read the next 4 bytes as the ID (32-bit integer)
          const id = view.getUint32(i * 5 + 1, false); // false for little-endian

          // Log the parsed data
          if (action === 1 || action === 0) {
            updatedState[id] = action === 1 ? true : false;
          } else {
            console.error(`Invalid action: ${action}`);
          }
          console.log("Action: ", action, " ID: ", id);
        }

        return updatedState;
      });
    })
    .catch((error) => {
      console.error("Failed to read Blob:", error);
    });
};
