import { useEffect, useState } from "react";

const query = "(prefers-color-scheme: dark)";

export default function useDarkMode() {
  const [dark, setDark] = useState(() => window.matchMedia(query).matches);

  useEffect(() => {
    const media = window.matchMedia(query);
    const listener = (event) => setDark(event.matches);

    media.addEventListener("change", listener);

    return () => media.removeEventListener("change", listener);
  }, []);

  return dark;
}
