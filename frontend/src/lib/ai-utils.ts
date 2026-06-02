export function getTextFromParts(parts: { type: string; text?: string }[]): string {
  return parts.filter((p) => p.type === "text").map((p) => p.text ?? "").join("");
}
