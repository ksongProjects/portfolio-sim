"use client";

import { useEffect, useState } from "react";

export type MarketStatus = "Pre-Market" | "Regular Hours" | "After Hours" | "Closed";

interface MarketState {
  status: MarketStatus;
  time: string;
}

interface EasternDateTimeParts {
  year: number;
  month: number;
  day: number;
  weekday: string;
  hours: number;
  minutes: number;
}

interface YearMonthDay {
  year: number;
  month: number;
  day: number;
}

const EASTERN_DATE_TIME_FORMATTER = new Intl.DateTimeFormat("en-US", {
  timeZone: "America/New_York",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  weekday: "short",
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
});

function formatEasternTime(date: Date): string {
  return date.toLocaleTimeString("en-US", {
    timeZone: "America/New_York",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  });
}

function getEasternDateTimeParts(date: Date): EasternDateTimeParts {
  const parts = EASTERN_DATE_TIME_FORMATTER.formatToParts(date);
  const value = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((part) => part.type === type)?.value ?? "";

  return {
    year: Number(value("year")),
    month: Number(value("month")),
    day: Number(value("day")),
    weekday: value("weekday"),
    hours: Number(value("hour")),
    minutes: Number(value("minute")),
  };
}

function utcDate(year: number, month: number, day: number) {
  return new Date(Date.UTC(year, month - 1, day));
}

function utcDateParts(date: Date): YearMonthDay {
  return {
    year: date.getUTCFullYear(),
    month: date.getUTCMonth() + 1,
    day: date.getUTCDate(),
  };
}

function addDays(date: YearMonthDay, days: number): YearMonthDay {
  return utcDateParts(utcDate(date.year, date.month, date.day + days));
}

function sameDate(left: YearMonthDay, right: YearMonthDay) {
  return left.year === right.year && left.month === right.month && left.day === right.day;
}

function observedFixedHoliday(year: number, month: number, day: number): YearMonthDay {
  const date = utcDate(year, month, day);
  const dayOfWeek = date.getUTCDay();

  if (dayOfWeek === 0) {
    return addDays({ year, month, day }, 1);
  }
  if (dayOfWeek === 6) {
    return addDays({ year, month, day }, -1);
  }
  return { year, month, day };
}

function nthWeekdayOfMonth(year: number, month: number, weekday: number, nth: number) {
  const firstDay = utcDate(year, month, 1).getUTCDay();
  const offset = (weekday - firstDay + 7) % 7;
  return 1 + offset + (nth - 1) * 7;
}

function lastWeekdayOfMonth(year: number, month: number, weekday: number) {
  const lastDate = utcDate(year, month + 1, 0);
  const offset = (lastDate.getUTCDay() - weekday + 7) % 7;
  return lastDate.getUTCDate() - offset;
}

function easterDate(year: number): YearMonthDay {
  const a = year % 19;
  const b = Math.floor(year / 100);
  const c = year % 100;
  const d = Math.floor(b / 4);
  const e = b % 4;
  const f = Math.floor((b + 8) / 25);
  const g = Math.floor((b - f + 1) / 3);
  const h = (19 * a + b - d - g + 15) % 30;
  const i = Math.floor(c / 4);
  const k = c % 4;
  const l = (32 + 2 * e + 2 * i - h - k) % 7;
  const m = Math.floor((a + 11 * h + 22 * l) / 451);
  const month = Math.floor((h + l - 7 * m + 114) / 31);
  const day = ((h + l - 7 * m + 114) % 31) + 1;

  return { year, month, day };
}

function marketHolidaysForYear(year: number): YearMonthDay[] {
  const easter = easterDate(year);

  return [
    observedFixedHoliday(year, 1, 1),
    { year, month: 1, day: nthWeekdayOfMonth(year, 1, 1, 3) },
    { year, month: 2, day: nthWeekdayOfMonth(year, 2, 1, 3) },
    addDays(easter, -2),
    { year, month: 5, day: lastWeekdayOfMonth(year, 5, 1) },
    observedFixedHoliday(year, 6, 19),
    observedFixedHoliday(year, 7, 4),
    { year, month: 9, day: nthWeekdayOfMonth(year, 9, 1, 1) },
    { year, month: 11, day: nthWeekdayOfMonth(year, 11, 4, 4) },
    observedFixedHoliday(year, 12, 25),
  ];
}

function isMarketHoliday(date: YearMonthDay) {
  return [date.year - 1, date.year, date.year + 1].some((year) =>
    marketHolidaysForYear(year).some((holiday) => sameDate(date, holiday))
  );
}

export function getMarketStatus(date: Date): MarketStatus {
  const { year, month, day, weekday, hours, minutes } = getEasternDateTimeParts(date);
  if (weekday === "Sat" || weekday === "Sun" || isMarketHoliday({ year, month, day })) {
    return "Closed";
  }

  const totalMinutes = hours * 60 + minutes;

  const preMarketStart = 4 * 60;
  const marketOpen = 9 * 60 + 30;
  const marketClose = 16 * 60;
  const afterHoursEnd = 20 * 60;

  if (totalMinutes >= preMarketStart && totalMinutes < marketOpen) {
    return "Pre-Market";
  }
  if (totalMinutes >= marketOpen && totalMinutes < marketClose) {
    return "Regular Hours";
  }
  if (totalMinutes >= marketClose && totalMinutes < afterHoursEnd) {
    return "After Hours";
  }
  return "Closed";
}

export function useMarketStatus() {
  const [marketState, setMarketState] = useState<MarketState>({
    status: "Closed",
    time: "--:--:--",
  });

  useEffect(() => {
    const update = () => {
      const now = new Date();
      setMarketState({
        status: getMarketStatus(now),
        time: formatEasternTime(now),
      });
    };

    update();
    const interval = setInterval(update, 1000);
    return () => clearInterval(interval);
  }, []);

  const isLive =
    marketState.status === "Pre-Market" ||
    marketState.status === "Regular Hours" ||
    marketState.status === "After Hours";

  return { status: marketState.status, time: marketState.time, isLive };
}
