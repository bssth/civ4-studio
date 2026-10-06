import {editor} from "../wailsjs/go/models";
import {i18n} from "./i18n";

/**
 * Text of a problem from the backend in the interface language. Problems have a code and arguments;
 * the English message is used when the code is not known to this version of the frontend.
 */
export function problemText(p: editor.Problem): string {
    const {t, te} = i18n.global;
    const key = p.code === 'unknownType' ? 'problems.unknownType' : `problems.${p.code}`;
    if (!p.code || !te(key)) return p.message;

    const args: Record<string, string> = {...(p.args ?? {})};
    if (p.code === 'unknownType' && te(`what.${args.what}`)) {
        args.what = t(`what.${args.what}`);
    }
    // Relation codes of the backend are the keys of relation names: contact, war, openBorders...
    if (p.code === 'teams.missingRelation' && te(`relations.${args.relation}`)) {
        args.relation = t(`relations.${args.relation}`);
    }
    let text = t(key, args);
    if (p.code === 'unknownType' && Number(args.count) > 1) {
        text += ' ' + t('problems.times', {n: args.count});
    }
    return text;
}

