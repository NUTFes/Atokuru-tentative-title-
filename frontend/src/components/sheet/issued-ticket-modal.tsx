"use client";

import { QRCodeSVG } from "qrcode.react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useT } from "@/lib/i18n";

export type IssuedTicketModalProps = {
  /** モーダルを開いているか */
  open: boolean;
  /** 開閉が変わったときの処理 */
  onOpenChange: (open: boolean) => void;
  /** 整理番号 */
  orderNumber: number;
  /** QR コードにする URL（id ではなく public_token を含めること） */
  qrValue: string;
};

/** 発券後に QR コードと整理番号を表示するモーダル。見た目だけを担当する */
export function IssuedTicketModal({
  open,
  onOpenChange,
  orderNumber,
  qrValue,
}: IssuedTicketModalProps) {
  const { t } = useT();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t.stall.issuedModal.title}</DialogTitle>
          <DialogDescription>{t.stall.issuedModal.guide}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col items-center gap-4 py-2">
          <div className="rounded-lg bg-white p-3">
            <QRCodeSVG value={qrValue} size={200} />
          </div>
          <div className="text-center">
            <p className="text-muted-foreground text-xs">{t.ticket.orderNumber}</p>
            <p className="text-5xl font-bold tabular-nums">{orderNumber}</p>
          </div>
        </div>
        <DialogFooter>
          <Button className="w-full" onClick={() => onOpenChange(false)}>
            {t.common.close}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
