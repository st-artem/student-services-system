import express, { Request, Response } from 'express';
import { PrismaClient } from '@prisma/client';
import dotenv from 'dotenv';

dotenv.config();

const prisma = new PrismaClient();
const app = express();
app.use(express.json());

// --- ДОПОМІЖНІ ФУНКЦІЇ ---
const sendError = (res: Response, code: number, message: string) => {
    res.status(code).json({ error: { code, message, timestamp: new Date().toISOString() } });
};

// --- ЕНДПОІНТИ ---

// 1. Створення гаманця (для тестів)
app.post('/wallets', async (req: Request, res: Response) => {
    const { currency } = req.body;
    if (!currency) return sendError(res, 400, "Currency is required");
    const wallet = await prisma.wallet.create({ data: { currency } });
    res.status(201).json(wallet);
});

// 2. Отримання балансу (ІЗ №8)
app.get('/wallets/:id/balance', async (req: Request, res: Response) => {
    const wallet = await prisma.wallet.findUnique({ where: { id: req.params.id } });
    if (!wallet) return sendError(res, 404, "Wallet not found");

    res.status(200).json({
        wallet_id: wallet.id,
        currency: wallet.currency,
        materialized_balance: wallet.balance.toNumber(),
        explanation: "This is the cached materialized balance for fast read operations, updated synchronously within transactions."
    });
});

// 3. Історія транзакцій (ІЗ №3)
app.get('/wallets/:id/history', async (req: Request, res: Response) => {
    const limit = parseInt(req.query.limit as string) || 10;
    const offset = parseInt(req.query.offset as string) || 0;

    const entries = await prisma.ledgerEntry.findMany({
        where: { walletId: req.params.id },
        take: limit,
        skip: offset,
        orderBy: { createdAt: 'desc' }
    });
    res.status(200).json(entries);
});

// 4. DEPOSIT (З ідемпотентністю ІЗ №2, 7)
app.post('/wallets/:id/deposit', async (req: Request, res: Response) => {
    const { amount, request_id } = req.body;
    if (!amount || amount <= 0 || !request_id) return sendError(res, 400, "Invalid amount or missing request_id");

    try {
        const result = await prisma.$transaction(async (tx) => {
            // ІЗ №6: Pessimistic Locking
            const wallet = await tx.$queryRaw<any[]>`SELECT * FROM wallets WHERE id = ${req.params.id}::uuid FOR UPDATE`;
            if (wallet.length === 0) throw new Error("NOT_FOUND");

            const updatedWallet = await tx.wallet.update({
                where: { id: req.params.id },
                data: { balance: { increment: amount } }
            });

            const entry = await tx.ledgerEntry.create({
                data: { walletId: req.params.id, operationType: "DEPOSIT", amount, requestId: request_id }
            });

            return { wallet: updatedWallet, entry };
        });

        res.status(200).json(result);
    } catch (error: any) {
        if (error.code === 'P2002') return sendError(res, 409, "Idempotency conflict: request_id already processed");
        if (error.message === "NOT_FOUND") return sendError(res, 404, "Wallet not found");
        sendError(res, 500, error.message);
    }
});

// 5. WITHDRAW (ІЗ №9)
app.post('/wallets/:id/withdraw', async (req: Request, res: Response) => {
    const { amount, request_id } = req.body;
    if (!amount || amount <= 0 || !request_id) return sendError(res, 400, "Invalid amount or missing request_id");

    try {
        const result = await prisma.$transaction(async (tx) => {
            // ІЗ №6: Pessimistic Locking
            const wallets = await tx.$queryRaw<any[]>`SELECT balance FROM wallets WHERE id = ${req.params.id}::uuid FOR UPDATE`;
            if (wallets.length === 0) throw new Error("NOT_FOUND");
            
            if (parseFloat(wallets[0].balance) < amount) throw new Error("INSUFFICIENT_FUNDS");

            const updatedWallet = await tx.wallet.update({
                where: { id: req.params.id },
                data: { balance: { decrement: amount } }
            });

            const entry = await tx.ledgerEntry.create({
                data: { walletId: req.params.id, operationType: "WITHDRAW", amount, requestId: request_id }
            });

            return { wallet: updatedWallet, entry };
        });

        res.status(200).json(result);
    } catch (error: any) {
        if (error.message === "INSUFFICIENT_FUNDS") return sendError(res, 422, "Insufficient funds");
        if (error.code === 'P2002') return sendError(res, 409, "Idempotency conflict: request_id already processed");
        if (error.message === "NOT_FOUND") return sendError(res, 404, "Wallet not found");
        sendError(res, 500, error.message);
    }
});

// 6. TRANSFER (Центральна операція)
app.post('/transfer', async (req: Request, res: Response) => {
    const { from_wallet_id, to_wallet_id, amount, request_id } = req.body;
    if (!from_wallet_id || !to_wallet_id || !amount || amount <= 0 || !request_id) {
        return sendError(res, 400, "Invalid input parameters");
    }

    try {
        const result = await prisma.$transaction(async (tx) => {
            // ІЗ №6: Запобігання Deadlocks (Сортування ID перед блокуванням)
            const ids = [from_wallet_id, to_wallet_id].sort();
            
            // Pessimistic Lock обох гаманців
            const wallets = await tx.$queryRaw<any[]>`SELECT id, balance, currency FROM wallets WHERE id IN (${ids[0]}::uuid, ${ids[1]}::uuid) FOR UPDATE`;
            
            if (wallets.length !== 2) throw new Error("NOT_FOUND");

            const wFrom = wallets.find(w => w.id === from_wallet_id);
            const wTo = wallets.find(w => w.id === to_wallet_id);

            // ІЗ №1: Перевірка валют
            if (wFrom.currency !== wTo.currency) throw new Error("CURRENCY_MISMATCH");
            
            // Перевірка балансу
            if (parseFloat(wFrom.balance) < amount) throw new Error("INSUFFICIENT_FUNDS");

            // Зміна балансів
            await tx.wallet.update({ where: { id: from_wallet_id }, data: { balance: { decrement: amount } } });
            await tx.wallet.update({ where: { id: to_wallet_id }, data: { balance: { increment: amount } } });

            // Створення операції та проводок
            const transfer = await tx.transferOperation.create({
                data: {
                    fromWalletId: from_wallet_id,
                    toWalletId: to_wallet_id,
                    amount,
                    currency: wFrom.currency,
                    requestId: request_id,
                    entries: {
                        create: [
                            { walletId: from_wallet_id, operationType: "WITHDRAW", amount, requestId: request_id + "_W" },
                            { walletId: to_wallet_id, operationType: "DEPOSIT", amount, requestId: request_id + "_D" }
                        ]
                    }
                }
            });

            return transfer;
        });

        res.status(200).json(result);
    } catch (error: any) {
        if (error.message === "INSUFFICIENT_FUNDS") return sendError(res, 422, "Insufficient funds");
        if (error.message === "CURRENCY_MISMATCH") return sendError(res, 400, "Currency mismatch between wallets");
        if (error.code === 'P2002') return sendError(res, 409, "Idempotency conflict: request_id already processed");
        if (error.message === "NOT_FOUND") return sendError(res, 404, "One or both wallets not found");
        sendError(res, 500, error.message);
    }
});

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
    console.log(`Wallet Service running on port ${PORT}`);
});