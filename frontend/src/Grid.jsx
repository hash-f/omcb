import React from "react";

import { FixedSizeGrid } from "react-window";

import { useWindowSize } from "./hooks/useWindowSize";
import { Cell } from "./Cell";

export const Grid = ({ checkHandler, bits }) => {
  const totalCount = 1_000_000;
  const cellHeight = 30;
  const cellWidth = 30;

  const { width, height } = useWindowSize();
  const gridHeight = height - 150;
  const gridWidth = width - 50;

  const columnCount = Math.floor(gridWidth / 30);
  const cellConfig = { columnCount, checkHandler, bits };

  return (
    <FixedSizeGrid
      columnCount={columnCount}
      rowCount={Math.ceil(totalCount / columnCount)}
      columnWidth={cellWidth}
      rowHeight={cellHeight}
      height={gridHeight}
      width={gridWidth}
      itemData={cellConfig}
    >
      {Cell}
    </FixedSizeGrid>
  );
};
