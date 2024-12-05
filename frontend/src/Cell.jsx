import React from "react";

export const Cell = ({ columnIndex, rowIndex, style, data }) => {
  const id = rowIndex * data.columnCount + columnIndex;
  if (id > 999_999) {
    return <></>;
  }
  const styles = {
    cell: {
      ...style,
      display: "inline-flex",
      alignItems: "center",
      justifyContent: "center",
    },
    checkbox: {
      width: 24,
      height: 24,
      cursor: "pointer",
      accentColor: "#097969",
    },
  };
  return (
    <div style={styles.cell}>
      <input
        type="checkbox"
        id={id}
        title={"Checkbox " + id}
        style={styles.checkbox}
        checked={data.bits[id]}
        onChange={data.checkHandler}
      />
    </div>
  );
};
