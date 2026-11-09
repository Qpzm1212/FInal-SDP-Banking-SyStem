// Запрос к серверу. Никогда не бросает исключение: всегда {ok, data}.
async function call(url, options) {
    try {
        const res = await fetch(url, options);
        let data = {};
        try { data = await res.json(); } catch (e) { /* пустой ответ */ }
        return { ok: res.ok, data };
    } catch (e) {
        return { ok: false, data: { error: 'Сервер недоступен' } };
    }
}

const post = (url, params) => call(url, { method: 'POST', body: new URLSearchParams(params) });

function showMessage(el, kind, text) {
    el.className = 'mt-3 fw-bold text-' + kind; // success | danger | warning
    el.textContent = text;
}

const errorText = (data) => 'Ошибка: ' + (data.error || 'неизвестная ошибка');
const value = (id) => document.getElementById(id).value.trim();

// 1. Транзакция (Observer)
const txForm = document.getElementById('txForm');
if (txForm) {
    txForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const out = document.getElementById('txResult');
        const { ok, data } = await post('/api/transaction', {
            account_id: value('accId'),
            amount: value('amount')
        });
        if (!ok) return showMessage(out, 'danger', errorText(data));
        showMessage(out, data.suspicious ? 'warning' : 'success', data.status);
    });
}

// 2. Открытие счета (Factory Method)
const accountForm = document.getElementById('accountForm');
if (accountForm) {
    accountForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const out = document.getElementById('accountResult');
        const { ok, data } = await post('/api/account', {
            type: value('newAccType'),
            id: value('newAccId')
        });
        if (!ok) return showMessage(out, 'danger', errorText(data));
        showMessage(out, 'success', data.message + ': ' + data.account_id + ' (' + data.account_type + ')');
    });
}

// 3. Кредит (Builder)
const loanForm = document.getElementById('loanForm');
if (loanForm) {
    loanForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const out = document.getElementById('loanResult');
        const { ok, data } = await post('/api/loan', {
            borrower: value('borrower'),
            amount: value('loanAmount')
        });
        out.style.display = 'block';
        out.textContent = ok ? JSON.stringify(data, null, 2) : errorText(data);
    });
}

// 4. Проценты (Strategy)
const interestForm = document.getElementById('interestForm');
if (interestForm) {
    interestForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const out = document.getElementById('interestResult');
        const query = new URLSearchParams({
            type: value('interestType'),
            balance: value('interestBalance')
        });
        const { ok, data } = await call('/api/interest?' + query);
        if (!ok) return showMessage(out, 'danger', errorText(data));
        showMessage(out, 'success', 'Проценты за год: $' + Number(data.interest).toFixed(2));
    });
}