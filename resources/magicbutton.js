// source: https://www.joshwcomeau.com/react/rainbow-button/

// blue (#0099ff)
// gold (#ffd700)
// cyan (#00f8ff)
const rainbowColors = [
  "hsl(204deg 100% 50%)",
  "hsl(190deg 100% 50%)",
  "hsl(160deg 80% 50%)",
  "hsl(90deg 70% 50%)",
  "hsl(51deg 100% 50%)",
  "hsl(51deg 100% 50%)",
  "hsl(51deg 100% 50%)",
  "hsl(90deg 70% 50%)",
  "hsl(160deg 80% 50%)",
  "hsl(190deg 100% 50%)",
  "hsl(204deg 100% 50%)",
];
const paletteSize = rainbowColors.length;

// Number of milliseconds for each update
const intervalDelay = 2000;

const colorNames = [
  "--magic-rainbow-color-0",
  "--magic-rainbow-color-1",
  "--magic-rainbow-color-2",
];

// Register properties
colorNames.forEach((name, index) => {
  CSS.registerProperty({
    name,
    syntax: "<color>",
    inherits: false,
    initialValue: rainbowColors[index],
  });
});

let buttonElems = document.querySelectorAll(".magic");

let cycleIndex = 8;

window.setInterval(() => {
  // Shift every color up by one position.
  //
  // `% paletteSize` is a handy trick to ensure
  // that values "wrap around"; if we've exceeded
  // the number of items in the array, it loops
  // back to 0.
  //
  //
  if (buttonElems.length === 0) {
    buttonElems = document.querySelectorAll(".magic");
  }

  buttonElems.forEach((b, index) => {
    const nextColors = [
      rainbowColors[(cycleIndex + 1 + index) % paletteSize],
      rainbowColors[(cycleIndex + 2 + index) % paletteSize],
      rainbowColors[(cycleIndex + 3 + index) % paletteSize],
    ];

    // Apply these new colors, update the DOM.
    colorNames.forEach((name, index) => {
      b.style.setProperty(name, nextColors[index]);
    });

    // increment the cycle count, so that we advance
    // the colors in the next loop.
    cycleIndex++;
  });
}, intervalDelay);
