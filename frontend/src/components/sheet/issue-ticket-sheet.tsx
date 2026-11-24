"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useT } from "@/lib/i18n";

export type IssueTicketSheetProps = {
  /** シートを開いているか */
  open: boolean;
  /** 開閉が変わったときの処理 */
  onOpenChange: (open: boolean) => void;
  /** 発券ボタンを押したときの処理。入力された備考を受け取る */
  onIssue: (note: string) => void;
  /** 発券中か（true のあいだはボタンを押せない） */
  issuing?: boolean;
  /** 備考の定型文。押すと備考に入る。定型文と同じ備考で発券すると、来場者の画面に訳が出る */
  templates?: string[];
};

/** 店舗画面で整理券を発行するボトムシート。見た目だけを担当する */
export function IssueTicketSheet({
  open,
  onOpenChange,
  onIssue,
  issuing = false,
  templates = [],
}: IssueTicketSheetProps) {
  const { t } = useT();
  const s = t.stall.issueSheet;

  // 入力中の文字だけは見た目の状態なのでここで持つ
  const [note, setNote] = useState("");

  // シートが閉じたときに備考を消す。発券に失敗してシートが開いたままなら、入力は残る
  const [prevOpen, setPrevOpen] = useState(open);
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (!open) setNote("");
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>{s.title}</DrawerTitle>
          <DrawerDescription>{s.description}</DrawerDescription>
        </DrawerHeader>
        {templates.length > 0 && (
          <div className="flex flex-col gap-2 px-4 pb-2">
            <p className="text-muted-foreground text-xs">{s.templates}</p>
            <div className="flex flex-wrap gap-2">
              {templates.map((template) => (
                <Button
                  key={template}
                  type="button"
                  variant={note === template ? "default" : "outline"}
                  size="sm"
                  aria-pressed={note === template}
                  onClick={() => setNote(template)}
                >
                  {template}
                </Button>
              ))}
            </div>
          </div>
        )}
        <div className="flex flex-col gap-2 px-4">
          <Label htmlFor="ticket-note">{s.noteLabel}</Label>
          <Textarea
            id="ticket-note"
            value={note}
            placeholder={s.notePlaceholder}
            onChange={(e) => setNote(e.target.value)}
          />
        </div>
        <DrawerFooter>
          <Button
            className="w-full"
            disabled={issuing}
            onClick={() => {
              onIssue(note.trim());
            }}
          >
            {issuing ? s.submitting : s.submit}
          </Button>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
