const FIRST_NAMES = [
  'Ana', 'Carlos', 'Maria', 'João', 'Fernanda', 'Pedro', 'Juliana', 'Lucas',
  'Beatriz', 'Rafael', 'Camila', 'Gustavo', 'Larissa', 'Thiago', 'Mariana',
  'Bruno', 'Isabela', 'Diego', 'Letícia', 'André', 'Patrícia', 'Rodrigo',
  'Vanessa', 'Felipe', 'Renata', 'Marcelo', 'Tatiana', 'Eduardo', 'Priscila',
  'Ricardo', 'Aline', 'Daniel', 'Natália', 'Fábio', 'Cláudia', 'Vinícius',
];

const LAST_NAMES = [
  'Silva', 'Santos', 'Oliveira', 'Souza', 'Rodrigues', 'Ferreira', 'Almeida',
  'Pereira', 'Lima', 'Gomes', 'Costa', 'Ribeiro', 'Martins', 'Carvalho',
  'Araújo', 'Melo', 'Barbosa', 'Rocha', 'Dias', 'Nascimento', 'Andrade',
  'Moreira', 'Nunes', 'Marques', 'Machado', 'Mendes', 'Freitas', 'Cardoso',
];

const SPECIALTIES = [
  { name: 'Cardiology', keywords: ['heart', 'cardiovascular', 'hypertension', 'arrhythmia'] },
  { name: 'Dermatology', keywords: ['skin', 'dermatitis', 'acne', 'melanoma'] },
  { name: 'Neurology', keywords: ['brain', 'nervous-system', 'epilepsy', 'migraine'] },
  { name: 'Orthopedics', keywords: ['bones', 'joints', 'fractures', 'spine'] },
  { name: 'Pediatrics', keywords: ['children', 'infant', 'vaccination', 'growth'] },
  { name: 'Psychiatry', keywords: ['mental-health', 'anxiety', 'depression', 'therapy'] },
  { name: 'Ophthalmology', keywords: ['eyes', 'vision', 'glaucoma', 'cataract'] },
  { name: 'Endocrinology', keywords: ['hormones', 'diabetes', 'thyroid', 'metabolism'] },
  { name: 'Gastroenterology', keywords: ['digestive', 'stomach', 'liver', 'intestine'] },
  { name: 'Pulmonology', keywords: ['lungs', 'respiratory', 'asthma', 'pneumonia'] },
  { name: 'Acupuncture', keywords: ['chinese-medicine', 'meridians', 'pain-relief', 'holistic'] },
  { name: 'Veterinary Medicine', keywords: ['animals', 'pets', 'surgery', 'vaccination'] },
  { name: 'Physiotherapy', keywords: ['rehabilitation', 'movement', 'posture', 'recovery'] },
  { name: 'Nutrition', keywords: ['diet', 'metabolism', 'weight', 'supplements'] },
];

const DESCRIPTIONS = [
  'Specialist with over 15 years of clinical experience in diagnosis and treatment',
  'Dedicated professional focused on evidence-based patient care',
  'Experienced practitioner combining traditional and modern approaches',
  'Board-certified specialist committed to personalized treatment plans',
  'Clinical researcher and practitioner with international training',
  'Holistic care provider with emphasis on preventive medicine',
  'Senior specialist with expertise in complex and rare conditions',
  'Multidisciplinary professional with hospital and private practice experience',
];

const LICENSE_PREFIXES = ['CRM-SP', 'CRM-RJ', 'CRM-MG', 'CRM-RS', 'CRM-PR', 'CRM-BA', 'CRMV-SP', 'CRMV-RJ'];

const PHONE_DDDS = ['11', '21', '31', '41', '51', '61', '71', '81', '85', '92'];

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function pickN(arr, n) {
  const shuffled = arr.slice().sort(() => Math.random() - 0.5);
  return shuffled.slice(0, n);
}

function uid() {
  return `${Date.now()}-${Math.floor(Math.random() * 1000000)}`;
}

function generatePhone() {
  const ddd = pick(PHONE_DDDS);
  const num = Math.floor(Math.random() * 900000000) + 100000000;
  return `+55${ddd}${num}`;
}

function generateLicense() {
  return `${pick(LICENSE_PREFIXES)}-${uid()}`;
}

function generateEmail(firstName, lastName) {
  const clean = (s) => s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
  return `${clean(firstName)}.${clean(lastName)}.${uid()}@healing-test.com`;
}

export function generateSpecialistData(overrides = {}) {
  const firstName = pick(FIRST_NAMES);
  const lastName = pick(LAST_NAMES);
  const specialty = pick(SPECIALTIES);

  const base = {
    name: `Dr. ${firstName} ${lastName}`,
    email: generateEmail(firstName, lastName),
    phone: generatePhone(),
    specialty: specialty.name,
    license_number: generateLicense(),
    description: `${pick(DESCRIPTIONS)} in ${specialty.name}`,
    keywords: pickN(specialty.keywords, 2 + Math.floor(Math.random() * 3)),
    agreed_to_share: true,
  };

  return Object.assign(base, overrides);
}