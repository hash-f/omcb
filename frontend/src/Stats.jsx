import React from "react";

export const Stats = ({ stats }) => {
  const formatWithCommas = (x) => {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  };

  return (
    <div>
      Toggled{" "}
      <span
        style={{
          borderRadius: "5px",
          border: "solid 1px #087364",
          backgroundColor: "#077969",
          padding: "0 5px",
          color: "#fff",
          fontSize: "16px",
        }}
      >
        {formatWithCommas(stats.total)}
      </span>{" "}
      times so far. A non conforming replica of the famous experiment by{" "}
      <a target="_blank" href="https://onemillioncheckboxes.com/">
        Nolen
      </a>
    </div>
  );
};
