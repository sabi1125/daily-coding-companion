function DateToString(date: string): string {
  const dateObj = new Date(date)
  const formattedDate = dateObj.toLocaleDateString("en-US", {
    month: "short", // "Aug"
    day: "numeric", // "21"
    year: "numeric", // "2026"
  });

  return formattedDate
}

export default DateToString
