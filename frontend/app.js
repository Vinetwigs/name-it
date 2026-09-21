const $ = s => document.querySelector(s);

const idle = $('#idle');
const working = $('#working');
const overlay = $('#overlay');
const namingMode = $('#naming-mode');
const customPatternWrap = $('#custom-pattern-wrap');
const customPattern = $('#custom-pattern');

let state = {
    active: false,
    root: '',
    filesWatched: 0,
    questionable: 0
};

function render() {
    idle.classList.toggle('hidden', state.active);
    working.classList.toggle('hidden', !state.active);

    $('#status').textContent = state.active
        ? 'WORKING'
        : 'OFFLINE';

    $('#status').className =
        'status ' + (state.active ? 'on' : 'off');

    $('#path').textContent = state.root || '—';
    $('#watched').textContent = state.filesWatched || 0;
    $('#questionable').textContent =
        state.questionable || 0;
}

async function refresh() {
    state = await window.go.main.App.GetState();

    render();

    renderCrimes(
        await window.go.main.App.GetCrimes()
    );
}

function renderCrimes(items) {
    $('#crimes').innerHTML =
        (items || [])
            .map(
                c => `
                    <div class="crime">
                        <div>
                            <div class="name">
                                ${escapeHtml(c.filename)}
                            </div>

                            <small>
                                ${new Date(
                                    c.timestamp
                                ).toLocaleTimeString()}
                            </small>
                        </div>

                        <div class="score">
                            ${c.score}
                        </div>
                    </div>
                `
            )
            .join('') ||
        '<div class="crime">No crimes yet. Suspiciously clean.</div>';
}

function escapeHtml(s) {
    return String(s).replace(
        /[&<>'"]/g,
        c =>
            ({
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                "'": '&#39;',
                '"': '&quot;'
            })[c]
    );
}

$('#start').onclick = async () => {
    const path =
        await window.go.main.App.ChooseFolder();

    if (path) {
        await window.go.main.App.StartWorking(path);
    }

    await refresh();
};

$('#stop').onclick = async () => {
    await window.go.main.App.StopWorking();

    await refresh();
};

$('#pause').onclick = async () => {
    await window.go.main.App.Pause(30);

    $('#pause').textContent = 'PAUSED 30 MIN';

    setTimeout(() => {
        $('#pause').textContent = 'PAUSE 30 MIN';
    }, 1800000);
};

document
    .querySelectorAll('.levels button')
    .forEach(button => {
        button.onclick = async () => {
            document
                .querySelectorAll('.levels button')
                .forEach(item => {
                    item.classList.remove('selected');
                });

            button.classList.add('selected');

            await window.go.main.App.SetAggressiveness(
                button.dataset.level
            );
        };
    });

$('#dismiss').onclick = async () => {
    overlay.classList.add('hidden');

    if (window.go?.main?.App?.DismissNotice) {
        await window.go.main.App.DismissNotice();
    }
};

window.runtime?.EventsOn?.(
    'nameit:crime',
    data => {
        state.filesWatched++;

        if (data.score >= 30) {
            state.questionable++;
        }

        render();

        showCrime(data);

        window.go.main.App
            .GetCrimes()
            .then(renderCrimes);
    }
);

function showCrime(data) {
    $('#overlay-message').textContent =
        data.message.split('\n\nFILE:')[0];

    $('#overlay-file').textContent =
        data.filename;

    overlay.classList.remove('hidden');

    document.body.animate(
        [
            { transform: 'translateX(-4px)' },
            { transform: 'translateX(4px)' },
            { transform: 'translateX(-2px)' },
            { transform: 'translateX(0)' }
        ],
        {
            duration: 120,
            iterations: 1
        }
    );
}

function updateCustomPatternVisibility() {
    customPatternWrap.classList.toggle(
        'hidden',
        namingMode.value !== 'custom'
    );
}

namingMode.onchange = async () => {
    const mode = namingMode.value;

    updateCustomPatternVisibility();

    await window.go.main.App.SetNamingMode(mode);
};

customPattern.oninput = async () => {
    await window.go.main.App.SetCustomPattern(
        customPattern.value
    );
};

updateCustomPatternVisibility();
refresh();
