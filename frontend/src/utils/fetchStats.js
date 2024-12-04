import { BASE_HTTP_URL } from "./constants.js";

// Fetch the binary data from the API
export const fetchStats = async (setStats) => {
  try {
    const response = await fetch(`${BASE_HTTP_URL}/stats`);
    const stats = await response.json();
    setStats((prevState) => {
      const updatedState = { ...prevState };

      updatedState.total = stats.total;
      return updatedState;
    });
  } catch (error) {
    console.error("Error fetching state:", error);
  }
};
