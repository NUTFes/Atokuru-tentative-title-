"use client";

import { Trash2Icon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useT } from "@/lib/i18n";

export type CommentTemplateItem = {
  id: number;
  comment: string;
};

export type CommentTemplateEditorProps = {
  /** 登録してある定型文 */
  templates: CommentTemplateItem[];
  /** 登録ボタンを押したときの処理。登録できたら true を返す（入力欄を空にする） */
  onAdd: (comment: string) => Promise<boolean>;
  /** 削除ボタンを押したときの処理 */
  onDelete: (id: number) => void;
  /** 登録・削除の途中か */
  busy?: boolean;
  /** 出すエラーの文。なければ null */
  error?: string | null;
};

/** 模擬店の「その他」で、備考の定型文を登録・削除するカード。見た目だけを担当する */
export function CommentTemplateEditor({
  templates,
  onAdd,
  onDelete,
  busy = false,
  error = null,
}: CommentTemplateEditorProps) {
  const { t } = useT();
  const s = t.stall.templates;
  // 入力中の文字だけは見た目の状態なのでここで持つ
  const [comment, setComment] = useState("");
  const trimmed = comment.trim();

  return (
    <Card>
      <CardHeader>
        <CardTitle>{s.title}</CardTitle>
        <CardDescription>{s.description}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (trimmed === "") return;
            void onAdd(trimmed).then((added) => {
              if (added) setComment("");
            });
          }}
        >
          <Input
            value={comment}
            maxLength={200}
            placeholder={s.placeholder}
            aria-label={s.title}
            onChange={(e) => setComment(e.target.value)}
          />
          <Button type="submit" disabled={busy || trimmed === ""}>
            {s.add}
          </Button>
        </form>
        {error && (
          <p role="alert" className="text-destructive text-sm">
            {error}
          </p>
        )}
        {templates.length === 0 ? (
          <p className="text-muted-foreground text-sm">{s.empty}</p>
        ) : (
          <ul className="flex flex-col divide-y">
            {templates.map((template) => (
              <li key={template.id} className="flex items-center gap-2 py-2">
                <span className="flex-1 text-sm break-words">{template.comment}</span>
                <Button
                  variant="ghost"
                  size="icon"
                  disabled={busy}
                  aria-label={`${s.delete}：${template.comment}`}
                  onClick={() => onDelete(template.id)}
                >
                  <Trash2Icon aria-hidden />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
