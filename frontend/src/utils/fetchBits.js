import { BASE_HTTP_URL } from "./constants.js";

// Fetch the binary data from the API
export const fetchBits = async (setBits) => {
  try {
    const response = await fetch(`${BASE_HTTP_URL}/state`);
    const data = await response.arrayBuffer();
    const byteArray = new Uint8Array(data);
    setBits((prevState) => {
      const updatedState = [...prevState];

      let i = 0;
      for (let byte of byteArray) {
        for (let j = 0; j < 8; j++) {
          updatedState[i] = ((byte >> (7 - j)) & 1) === 1;
          i++;
        }
      }
      return updatedState;
    });
  } catch (error) {
    console.error("Error fetching state:", error);
  }
};
